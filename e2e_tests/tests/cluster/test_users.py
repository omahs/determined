import contextlib
import json
import logging
import os
import pathlib
import re
import shutil
import time
import uuid
from typing import Dict, Generator, Iterator, List, Optional, Tuple, cast

import appdirs
import pexpect
import pytest
from pexpect import spawn

from determined.common import api, constants, util
from determined.common.api import authentication, bindings, errors
from determined.experimental import Determined
from tests import api_utils, command
from tests import config as conf
from tests import detproc
from tests import experiment as exp
from tests.filetree import FileTree

EXPECT_TIMEOUT = 5
logger = logging.getLogger(__name__)


@pytest.fixture()
def clean_auth() -> Iterator[None]:
    """
    clean_auth is a session-level fixture that ensures that we run tests with no preconfigured
    authentication, and that any settings we save during tests are cleaned up afterwards.
    """
    authentication.TokenStore(conf.make_master_url()).delete_token_cache()
    yield None
    authentication.TokenStore(conf.make_master_url()).delete_token_cache()


def get_random_string() -> str:
    return str(uuid.uuid4())


def det_run(sess: api.Session, args: List[str]) -> str:
    return detproc.check_output(sess, ["det", *args]).decode()


def activate_deactivate_user(sess: api.Session, active: bool, target_user: str) -> None:
    command = [
        "det",
        "user",
        "activate" if active else "deactivate",
        target_user,
    ]
    detproc.check_output(sess, command)


def extract_columns(output: str, column_indices: List[int]) -> List[Tuple[str, ...]]:
    lines = output.split("\n")
    # Ignore the header.
    lines = lines[2:]
    parsed = []
    for line in lines:
        if not line:
            continue
        columns = line.split("|")
        parsed.append(tuple(columns[i].strip() for i in column_indices))

    return parsed


def extract_id_and_owner_from_exp_list(output: str) -> List[Tuple[int, str]]:
    rows = extract_columns(output, [0, 1])
    return [(int(r[0]), r[1]) for r in rows]


@pytest.mark.e2e_cpu
def test_post_user_api(clean_auth: None, login_admin: None) -> None:
    new_username = get_random_string()

    sess = api_utils.admin_session()

    user = bindings.v1User(active=True, admin=False, username=new_username)
    body = bindings.v1PostUserRequest(password="", user=user)
    resp = bindings.post_PostUser(sess, body=body)
    assert resp and resp.user
    assert resp.to_json()["user"]["username"] == new_username
    assert resp.user.agentUserGroup is None

    user = bindings.v1User(
        active=True,
        admin=False,
        username=get_random_string(),
        agentUserGroup=bindings.v1AgentUserGroup(
            agentUid=1000, agentGid=1001, agentUser="username", agentGroup="groupname"
        ),
    )
    resp = bindings.post_PostUser(sess, body=bindings.v1PostUserRequest(user=user))
    assert resp and resp.user and resp.user.agentUserGroup
    assert resp.user.agentUserGroup.agentUser == "username"
    assert resp.user.agentUserGroup.agentGroup == "groupname"
    assert resp.user.agentUserGroup.agentUid == 1000
    assert resp.user.agentUserGroup.agentGid == 1001

    user = bindings.v1User(
        active=True,
        admin=False,
        username=get_random_string(),
        agentUserGroup=bindings.v1AgentUserGroup(
            agentUid=1000,
            agentGid=1001,
        ),
    )

    with pytest.raises(errors.APIException):
        bindings.post_PostUser(sess, body=bindings.v1PostUserRequest(user=user))


@pytest.mark.e2e_cpu
def test_create_user_sdk(clean_auth: None, login_admin: None) -> None:
    username = get_random_string()
    password = get_random_string()
    det_obj = Determined(master=conf.make_master_url())
    user = det_obj.create_user(username=username, admin=False, password=password)
    assert user.user_id is not None and user.username == username


# XXX: fixme: split
@pytest.mark.e2e_cpu
def test_logout(clean_auth: None, login_admin: None) -> None:
    # Tests fallback to default determined user
    creds = api_utils.create_test_user(True)

    # Set Determined password to something in order to disable auto-login.
    password = get_random_string()
    assert change_user_password(constants.DEFAULT_DETERMINED_USER, password) == 0

    # Log in as new user.
    api_utils.configure_token_store(creds)
    # Now we should be able to list experiments.
    child = det_spawn(["e", "list"])
    child.read()
    child.wait()
    child.close()
    assert child.status == 0

    # Exiting the logged_in_user context logs out and asserts that the exit code is 0.
    log_out_user()
    # Now trying to list experiments should result in an error.
    child = det_spawn(["e", "list"])
    expected = "Unauthenticated"
    assert expected in str(child.read())
    child.wait()
    child.close()
    assert child.status != 0

    # Log in as determined.
    api_utils.configure_token_store(
        authentication.Credentials(constants.DEFAULT_DETERMINED_USER, password)
    )

    # Log back in as new user.
    api_utils.configure_token_store(creds)

    # Now log out as determined.
    log_out_user(constants.DEFAULT_DETERMINED_USER)

    # Should still be able to list experiments because new user is logged in.
    child = det_spawn(["e", "list"])
    child.read()
    child.wait()
    child.close()
    assert child.status == 0

    # Change Determined password back to "".
    change_user_password(constants.DEFAULT_DETERMINED_USER, "")
    # Clean up.


@pytest.mark.e2e_cpu
@pytest.mark.e2e_cpu_postgres
def test_activate_deactivate(clean_auth: None, login_admin: None) -> None:
    sess, username, password = api_utils.create_test_user(True)

    # Deactivate user.
    admin = api_utils.admin_session()
    activate_deactivate_user(admin, False, username)

    # Attempt to log in again.
    with pytest.raises(api.UnauthorizedError):
        api_utils.make_session(username, password)

    # Activate user.
    activate_deactivate_user(admin, True, username)

    # Now log in again. It should have a non-zero exit status.
    api_utils.make_session(username, password)

    # SDK testing for activating and deactivating.
    det_obj = Determined._from_session(admin)
    user = det_obj.get_user_by_name(user_name=username)
    user.deactivate()
    assert user.active is not True
    with pytest.raises(api.UnauthorizedError):
        api_utils.make_session(username, password)

    user.activate()
    assert user.active is True
    api_utils.make_session(username, password)


# XXX: fixme; split
@pytest.mark.e2e_cpu
@pytest.mark.e2e_cpu_postgres
def test_change_password(clean_auth: None, login_admin: None) -> None:
    # Create a user without a password.
    creds = api_utils.create_test_user(False)

    # Attempt to log in.
    api_utils.configure_token_store(creds)

    # Log out.
    log_out_user()

    # login admin
    api_utils.configure_token_store(conf.ADMIN_CREDENTIALS)

    new_password = get_random_string()
    assert change_user_password(creds.username, new_password) == 0
    api_utils.configure_token_store(authentication.Credentials(creds.username, new_password))

    new_password_sdk = get_random_string()
    det_obj = Determined(master=conf.make_master_url())
    user = det_obj.get_user_by_name(user_name=creds.username)
    user.change_password(new_password=new_password_sdk)
    api_utils.configure_token_store(authentication.Credentials(creds.username, new_password_sdk))


# XXX: fixme; pure sdk
@pytest.mark.e2e_cpu
def test_change_own_password(clean_auth: None, login_admin: None) -> None:
    # Create a user without a password.
    creds = api_utils.create_test_user(False)
    log_in_user_cli(creds)
    assert change_user_password(creds.username, creds.password, own=True) == 0


@pytest.mark.e2e_cpu
def test_change_username(clean_auth: None, login_admin: None) -> None:
    creds = api_utils.create_test_user()
    new_username = "rename-user-64"
    command = ["det", "user", "rename", creds.username, new_username]
    detproc.run(command, check=True)
    det_obj = Determined(master=conf.make_master_url())
    user = det_obj.get_user_by_name(user_name=new_username)
    assert user.username == new_username
    api_utils.configure_token_store(authentication.Credentials(new_username, ""))

    # Test SDK
    new_username = "rename-user-$64"
    user.rename(new_username)
    user = det_obj.get_user_by_name(user_name=new_username)
    assert user.username == new_username
    api_utils.configure_token_store(authentication.Credentials(new_username, ""))


@pytest.mark.e2e_cpu
@pytest.mark.e2e_cpu_postgres
@pytest.mark.e2e_cpu_cross_version
def test_experiment_creation_and_listing(clean_auth: None, login_admin: None) -> None:
    # Create 2 users.
    sess1, user1, _ = api_utils.create_test_user(True)
    sess2, user2, _ = api_utils.create_test_user(True)

    # Create an experiment as first user.
    experiment_id1 = exp.run_basic_test(
        sess1, conf.fixtures_path("no_op/single.yaml"), conf.fixtures_path("no_op"), 1
    )

    # Create another experiment, this time as second user.
    experiment_id2 = exp.run_basic_test(
        sess2, conf.fixtures_path("no_op/single.yaml"), conf.fixtures_path("no_op"), 1
    )

    # user 1 can only see user 1 experiment
    output = extract_id_and_owner_from_exp_list(det_run(sess1, ["e", "list"]))
    assert (experiment_id1, user1) in output, output
    assert (experiment_id2, user2) not in output, output

    # Now use the -a flag to list all experiments.  The output should include both experiments.
    output = extract_id_and_owner_from_exp_list(det_run(sess1, ["e", "list", "-a"]))
    assert (experiment_id1, user1) in output, output
    assert (experiment_id2, user2) in output, output

    # Clean up.
    # XXX: why bother?  Seems to just cost extra time.
    delete_experiments(api_utils.admin_session(), experiment_id1, experiment_id2)


@pytest.mark.e2e_cpu
def test_login_wrong_password(clean_auth: None, login_admin: None) -> None:
    sess, username, password = api_utils.create_test_user(True)
    with pytest.raises(api.UnauthorizedError):
        api_utils.make_session(username, "wrong" + password)


@pytest.mark.e2e_cpu
def test_login_as_non_existent_user(clean_auth: None, login_admin: None) -> None:
    with pytest.raises(api.UnauthorizedError):
        api_utils.make_session("nOtArEaLuSeR", "password")


# XXX: This is straight-up a cli test
@pytest.mark.e2e_cpu
def test_login_with_environment_variables(clean_auth: None, login_admin: None) -> None:
    creds = api_utils.create_test_user(True)
    # logout admin
    log_out_user()

    os.environ["DET_USER"] = creds.username
    os.environ["DET_PASS"] = creds.password
    try:
        child = det_spawn(["user", "whoami"])
        child.expect(creds.username)
        child.read()
        child.wait()
        assert child.exitstatus == 0

        # Can still override with -u.
        with logged_in_user(conf.ADMIN_CREDENTIALS):
            child = det_spawn(["-u", conf.ADMIN_CREDENTIALS.username, "user", "whoami"])
            child.expect(conf.ADMIN_CREDENTIALS.username)
            child.read()
            child.wait()
            assert child.exitstatus == 0
    finally:
        del os.environ["DET_USER"]
        del os.environ["DET_PASS"]


@pytest.mark.e2e_cpu
def test_auth_inside_shell(clean_auth: None, login_admin: None) -> None:
    creds = api_utils.create_test_user(True)

    with logged_in_user(creds):
        # start a shell
        child = det_spawn(["shell", "start"])
        child.setecho(True)
        # shells take time to start; use the default timeout which is longer
        child.expect(r".*Permanently added.+([0-9a-f-]{36}).+known hosts\.")

        shell_id = child.match.group(1).decode("utf-8")

        def check_whoami(expected_username: str) -> None:
            child.sendline("det user whoami")
            child.expect("You are logged in as user \\'(.*)\\'", timeout=EXPECT_TIMEOUT)
            username = child.match.group(1).decode("utf-8")
            logger.debug(f"They are logged in as user {username}")
            assert username == expected_username

        # check the current user
        check_whoami(creds.username)

        # log in as admin
        child.sendline(f"det user login {conf.ADMIN_CREDENTIALS.username}")
        child.expect(
            f"Password for user '{conf.ADMIN_CREDENTIALS.username}'", timeout=EXPECT_TIMEOUT
        )
        child.sendline(conf.ADMIN_CREDENTIALS.password)

        # check that whoami responds with the new user
        check_whoami(conf.ADMIN_CREDENTIALS.username)

        # log out
        child.sendline("det user logout")
        child.expect("#", timeout=EXPECT_TIMEOUT)

        # check that we are back to who we were
        check_whoami(creds.username)

        child.sendline("exit")

        child = det_spawn(["shell", "kill", shell_id])
        child.read()
        child.wait()
        assert child.exitstatus == 0


# XXX: delete this
@pytest.mark.e2e_cpu
def test_login_as_non_active_user(clean_auth: None, login_admin: None) -> None:
    creds = api_utils.create_test_user(True)

    passwd_prompt = f"Password for user '{creds.username}':"
    unauth_error = "user is not active"
    command = ["det", "user", "deactivate", creds.username]
    detproc.run(command, check=True)

    child = det_spawn(["user", "login", creds.username])
    child.setecho(True)
    child.expect(passwd_prompt, timeout=EXPECT_TIMEOUT)
    child.sendline(creds.password)
    assert unauth_error in str(child.read())
    child.wait()
    child.close()

    assert child.exitstatus != 0


@pytest.mark.e2e_cpu
def test_non_admin_user_link_with_agent_user(clean_auth: None, login_admin: None) -> None:
    sess = api_utils.user_session()
    _, username, _ = api_utils.create_test_user(True)

    cmd = [
        "det",
        "user",
        "link-with-agent-user",
        username,
        "--agent-uid",
        "1",
        "--agent-gid",
        "1",
        "--agent-user",
        username,
        "--agent-group",
        username,
    ]

    detproc.check_error(sess, cmd, "forbidden")


@pytest.mark.e2e_cpu
def test_non_admin_commands(clean_auth: None, login_admin: None) -> None:
    sess = api_utils.user_session()
    command = [
        "det",
        "slot",
        "list",
        "--json",
    ]
    slots = detproc.check_json(sess, command)

    slots = json.loads(output)
    slot_id = slots[0]["slot_id"]
    agent_id = slots[0]["agent_id"]

    enable_slots = ["slot", "enable", agent_id, slot_id]
    disable_slots = ["slot", "disable", agent_id, slot_id]
    enable_agents = ["agent", "enable", agent_id]
    disable_agents = ["agent", "disable", agent_id]
    config = ["master", "config"]
    for cmd in [disable_slots, disable_agents, enable_slots, enable_agents, config]:
        detproc.check_error(sess, ["det", *cmd], "forbidden")


def run_command(session: api.Session) -> str:
    body = bindings.v1LaunchCommandRequest(config={"entrypoint": ["echo", "hello"]})
    cmd = bindings.post_LaunchCommand(session, body=body).command
    return cmd.id


def start_notebook(sess: api.Session) -> str:
    return detproc.check_output(sess, ["det", "notebook", "start", "-d"]).strip()


def start_tensorboard(sess: api.Session, experiment_id: int) -> str:
    cmd = ["det", "tensorboard", "start", "-d", str(experiment_id)]
    return detproc.check_output(sess, cmd).strip()


def delete_experiments(sess: api.Session, *experiment_ids: int) -> None:
    eids = set(experiment_ids)
    while eids:
        output = extract_columns(det_run(sess, ["e", "list", "-a"]), [0, 4])

        running_ids = {int(o[0]) for o in output if o[1] == "COMPLETED"}
        intersection = eids & running_ids
        if not intersection:
            time.sleep(0.5)
            continue

        experiment_id = intersection.pop()
        det_run(sess, ["e", "delete", "--yes", str(experiment_id)])
        eids.remove(experiment_id)


def kill_notebooks(sess: api.Session, *notebook_ids: str) -> None:
    nids = set(notebook_ids)
    while nids:
        output = extract_columns(det_run(sess, ["notebook", "list", "-a"]), [0, 3])  # id, state

        # Get set of running IDs.
        running_ids = {task_id for task_id, state in output if state == "RUNNING"}

        intersection = running_ids & nids
        if not intersection:
            time.sleep(0.5)
            continue

        notebook_id = intersection.pop()
        detproc.check_output(sess, ["det", "notebook", "kill", notebook_id])
        nids.remove(notebook_id)


def kill_tensorboards(sess: api.Session, *tensorboard_ids: str) -> None:
    tids = set(tensorboard_ids)
    while tids:
        output = extract_columns(det_run(["tensorboard", "list", "-a"]), [0, 3])

        running_ids = {task_id for task_id, state in output if state == "RUNNING"}

        intersection = running_ids & tids
        if not intersection:
            time.sleep(0.5)
            continue

        tensorboard_id = intersection.pop()
        detproc.check_output(sess, ["det", "tensorboard", "kill", notebook_id])
        tids.remove(tensorboard_id)


@pytest.mark.e2e_cpu
def test_notebook_creation_and_listing(clean_auth: None, login_admin: None) -> None:
    creds1 = api_utils.create_test_user(True)
    creds2 = api_utils.create_test_user(True)

    with logged_in_user(creds1):
        notebook_id1 = start_notebook()

    with logged_in_user(creds2):
        notebook_id2 = start_notebook()

        # Listing should only give us user 2's experiment.
        output = extract_columns(det_run(["notebook", "list"]), [0, 1])

    with logged_in_user(creds1):
        output = extract_columns(det_run(["notebook", "list"]), [0, 1])
        assert (notebook_id1, creds1.username) in output
        assert (notebook_id2, creds2.username) not in output

        # Now test listing all.
        output = extract_columns(det_run(["notebook", "list", "-a"]), [0, 1])
        assert (notebook_id1, creds1.username) in output
        assert (notebook_id2, creds2.username) in output

    # Clean up, killing experiments.
    kill_notebooks(notebook_id1, notebook_id2)


@pytest.mark.e2e_cpu
def test_tensorboard_creation_and_listing(clean_auth: None, login_admin: None) -> None:
    creds1 = api_utils.create_test_user(True)
    creds2 = api_utils.create_test_user(True)

    with logged_in_user(creds1):
        # Create an experiment.
        experiment_id1 = exp.run_basic_test(
            conf.fixtures_path("no_op/single.yaml"), conf.fixtures_path("no_op"), 1
        )

    with logged_in_user(creds1):
        tensorboard_id1 = start_tensorboard(experiment_id1)

    with logged_in_user(creds2):
        experiment_id2 = exp.run_basic_test(
            conf.fixtures_path("no_op/single.yaml"), conf.fixtures_path("no_op"), 1
        )

    with logged_in_user(creds2):
        tensorboard_id2 = start_tensorboard(experiment_id2)

    with logged_in_user(creds1):
        output = extract_columns(det_run(["tensorboard", "list"]), [0, 1])
        assert (tensorboard_id1, creds1.username) in output
        assert (tensorboard_id2, creds2.username) not in output

        output = extract_columns(det_run(["tensorboard", "list", "-a"]), [0, 1])
        assert (tensorboard_id1, creds1.username) in output
        assert (tensorboard_id2, creds2.username) in output

    kill_tensorboards(tensorboard_id1, tensorboard_id2)

    with logged_in_user(conf.ADMIN_CREDENTIALS):
        delete_experiments(experiment_id1, experiment_id2)


@pytest.mark.e2e_cpu
def test_command_creation_and_listing(clean_auth: None) -> None:
    creds1 = api_utils.create_test_user(True)
    creds2 = api_utils.create_test_user(True)
    session1 = api_utils.make_session(creds1)
    session2 = api_utils.make_session(creds2)

    command_id1 = run_command(session=session1)

    command_id2 = run_command(session=session2)

    cmds = bindings.get_GetCommands(session1, users=[creds1.username]).commands
    output = [(cmd.id, cmd.username) for cmd in cmds]
    assert (command_id1, creds1.username) in output
    assert (command_id2, creds2.username) not in output

    cmds = bindings.get_GetCommands(session1).commands
    output = [(cmd.id, cmd.username) for cmd in cmds]
    assert (command_id1, creds1.username) in output
    assert (command_id2, creds2.username) in output


def create_linked_user(
    sess: api.Session, uid: int, user: str, gid: int, group: str
) -> authentication.Credentials:
    admin = api_utils.admin_session()
    sess, username, password = api_utils.create_test_user(False)

    cmd = [
        "det",
        "user",
        "link-with-agent-user",
        user_creds.username,
        "--agent-uid",
        str(uid),
        "--agent-gid",
        str(gid),
        "--agent-user",
        user,
        "--agent-group",
        group,
    ]

    detproc.check_output(admin, cmd)

    return sess, username, password


def create_linked_user_sdk(
    uid: int, agent_user: str, gid: int, group: str
) -> authentication.Credentials:
    creds = api_utils.create_test_user(False)
    det_obj = Determined(master=conf.make_master_url())
    user = det_obj.get_user_by_name(user_name=creds.username)
    user.link_with_agent(agent_gid=gid, agent_uid=uid, agent_group=group, agent_user=agent_user)
    return creds


def check_link_with_agent_output(user: authentication.Credentials, expected_output: str) -> None:
    with logged_in_user(user), command.interactive_command(
        "cmd", "run", "bash", "-c", "echo $(id -u -n):$(id -u):$(id -g -n):$(id -g)"
    ) as cmd:
        for line in cmd.stdout:
            if expected_output in line:
                break
        else:
            raise AssertionError(f"Did not find {expected_output} in output")


@pytest.mark.e2e_cpu
def test_link_with_agent_user(clean_auth: None, login_admin: None) -> None:
    user = create_linked_user(200, "someuser", 300, "somegroup")
    expected_output = "someuser:200:somegroup:300"
    check_link_with_agent_output(user, expected_output)

    with logged_in_user(conf.ADMIN_CREDENTIALS):
        user_sdk = create_linked_user_sdk(210, "anyuser", 310, "anygroup")
        expected_output = "anyuser:210:anygroup:310"
        check_link_with_agent_output(user_sdk, expected_output)


@pytest.mark.e2e_cpu
def test_link_with_large_uid(clean_auth: None, login_admin: None) -> None:
    user = create_linked_user(2000000000, "someuser", 2000000000, "somegroup")

    expected_output = "someuser:2000000000:somegroup:2000000000"
    check_link_with_agent_output(user, expected_output)


@pytest.mark.e2e_cpu
def test_link_with_existing_agent_user(clean_auth: None, login_admin: None) -> None:
    user = create_linked_user(65534, "nobody", 65534, "nogroup")

    expected_output = "nobody:65534:nogroup:65534"
    check_link_with_agent_output(user, expected_output)


@contextlib.contextmanager
def non_tmp_shared_fs_path() -> Generator:
    """
    Proper checkpoint storage handling for shared_fs involves properly choosing to use the
    container_path instead of the host_path. Issues don't really arise if the container is running
    as root (because root can write to anywhere) or if host_path is in /tmp (because /tmp is world
    writable) so this context manager yields a checkpoint storage config where host_path is a
    user-owned directory.

    Making it a user-owned directory ensures that the test runs without root privileges on
    normal developer machines, and it also ensures that the test would fail if the code was broken.

    Tests should not pollute user directories though, so make sure to clean up the checkpoint
    directory that we use.
    """

    cache_dir = appdirs.user_cache_dir("determined", "determined")
    checkpoint_dir = os.path.join(cache_dir, "e2e_tests")
    os.makedirs(checkpoint_dir, exist_ok=True)
    os.chmod(checkpoint_dir, 0o777)

    try:
        yield checkpoint_dir
    finally:
        shutil.rmtree(checkpoint_dir)


@pytest.mark.e2e_cpu
def test_non_root_experiment(clean_auth: None, login_admin: None, tmp_path: pathlib.Path) -> None:
    user = create_linked_user(65534, "nobody", 65534, "nogroup")

    with logged_in_user(user):
        with open(conf.fixtures_path("no_op/model_def.py")) as f:
            model_def_content = f.read()

        with open(conf.fixtures_path("no_op/single-one-short-step.yaml")) as f:
            config = util.yaml_safe_load(f)

        # Use a user-owned path to ensure shared_fs uses the container_path and not host_path.
        with non_tmp_shared_fs_path() as host_path:
            config["checkpoint_storage"] = {
                "type": "shared_fs",
                "host_path": host_path,
            }

            # Call `det --version` in a startup hook to ensure that det is on the PATH.
            with FileTree(
                tmp_path,
                {
                    "startup-hook.sh": "det --version || exit 77",
                    "const.yaml": util.yaml_safe_dump(config),
                    "model_def.py": model_def_content,
                },
            ) as tree:
                exp.run_basic_test(str(tree.joinpath("const.yaml")), str(tree), None)


@pytest.mark.e2e_cpu
def test_link_without_agent_user(clean_auth: None, login_admin: None) -> None:
    user = api_utils.create_test_user(False)

    expected_output = "root:0:root:0"
    with logged_in_user(user), command.interactive_command(
        "cmd", "run", "bash", "-c", "echo $(id -u -n):$(id -u):$(id -g -n):$(id -g)"
    ) as cmd:
        recvd = []
        for line in cmd.stdout:
            if expected_output in line:
                break
            recvd.append(line)
        else:
            output = "".join(recvd)
            raise AssertionError(f"Did not find {expected_output} in output:\n{output}")


@pytest.mark.e2e_cpu
def test_non_root_shell(clean_auth: None, login_admin: None, tmp_path: pathlib.Path) -> None:
    user = create_linked_user(1234, "someuser", 1234, "somegroup")

    expected_output = "someuser:1234:somegroup:1234"

    with logged_in_user(user), command.interactive_command("shell", "start") as shell:
        shell.stdin.write(b"echo $(id -u -n):$(id -u):$(id -g -n):$(id -g)\n")
        shell.stdin.close()

        for line in shell.stdout:
            if expected_output in line:
                break
        else:
            raise AssertionError(f"Did not find {expected_output} in output")


@pytest.mark.e2e_cpu
def test_experiment_delete(clean_auth: None, login_admin: None) -> None:
    sess = api_utils.user_session()
    other, _, _ = api_utils.create_test_user()

    experiment_id = exp.run_basic_test(
        sess, conf.fixtures_path("no_op/single.yaml"), conf.fixtures_path("no_op"), 1
    )

    # "det experiment delete" call should fail, because the other user is not an admin and
    # doesn't own the experiment.
    cmd = ["det", "experiment", "delete", str(experiment_id), "--yes"]
    detproc.check_error(other, cmd, "forbidden")

    # but the owner can delete it
    detproc.check_output(sess, cmd)

    experiment_delete_deadline = time.time() + 5 * 60
    while True:
        # "det experiment describe" call should fail, because the
        # experiment is no longer in the database.
        p = detproc.run(sess, ["det", "experiment", "describe", str(experiment_id)])
        if p.exitstatus != 0:
            return
        elif time.time() > experiment_delete_deadline:
            pytest.fail("experiment didn't delete after timeout")


def _fetch_user_by_username(sess: api.Session, username: str) -> bindings.v1User:
    # Get API bindings object for the created test user
    all_users = bindings.get_GetUsers(sess).users
    assert all_users is not None
    return next(u for u in all_users if u.username == username)


@pytest.mark.e2e_cpu
@pytest.mark.e2e_cpu_postgres
def test_change_displayname(clean_auth: None, login_admin: None) -> None:
    sess, original_name, _ = api_utils.create_test_user(False)

    current_user = _fetch_user_by_username(sess, original_name)
    assert current_user is not None and current_user.id

    # Rename user using display name
    patch_user = bindings.v1PatchUser(displayName="renamed display-name")
    bindings.patch_PatchUser(sess, body=patch_user, userId=current_user.id)

    modded_user = bindings.get_GetUser(sess, userId=current_user.id).user
    assert modded_user is not None
    assert modded_user.displayName == "renamed display-name"

    # Rename user display name using SDK
    det_obj = Determined(master=conf.make_master_url())
    user = det_obj.get_user_by_id(user_id=current_user.id)
    user.change_display_name(display_name="renamedSDK")

    modded_user_sdk = det_obj.get_user_by_id(user_id=current_user.id)
    assert modded_user_sdk is not None
    assert modded_user_sdk.display_name == "renamedSDK"

    # Avoid display name of 'admin'
    patch_user.displayName = "Admin"
    with pytest.raises(errors.APIException):
        bindings.patch_PatchUser(sess, body=patch_user, userId=current_user.id)

    # Clear display name (UI will show username)
    patch_user.displayName = ""
    bindings.patch_PatchUser(sess, body=patch_user, userId=current_user.id)

    modded_user = bindings.get_GetUser(sess, userId=current_user.id).user
    assert modded_user is not None
    assert modded_user.displayName == ""


@pytest.mark.e2e_cpu
def test_patch_agentusergroup(clean_auth: None, login_admin: None) -> None:
    test_user_credentials = api_utils.create_test_user(False)
    test_username = test_user_credentials.username

    # Patch - normal.
    sess = api_utils.admin_session()
    patch_user = bindings.v1PatchUser(
        agentUserGroup=bindings.v1AgentUserGroup(
            agentGid=1000, agentUid=1000, agentUser="username", agentGroup="groupname"
        )
    )
    test_user = _fetch_user_by_username(sess, test_username)
    assert test_user.id
    bindings.patch_PatchUser(sess, body=patch_user, userId=test_user.id)
    patched_user = bindings.get_GetUser(sess, userId=test_user.id).user
    assert patched_user is not None and patched_user.agentUserGroup is not None
    assert patched_user.agentUserGroup.agentUser == "username"
    assert patched_user.agentUserGroup.agentGroup == "groupname"

    # Patch - missing username/groupname.
    patch_user = bindings.v1PatchUser(
        agentUserGroup=bindings.v1AgentUserGroup(agentGid=1000, agentUid=1000)
    )
    test_user = _fetch_user_by_username(sess, test_username)
    assert test_user.id
    with pytest.raises(errors.APIException):
        bindings.patch_PatchUser(sess, body=patch_user, userId=test_user.id)


# XXX this should be a CLI test
@pytest.mark.e2e_cpu
def test_logout_all(clean_auth: None, login_admin: None) -> None:
    creds = api_utils.create_test_user(True)

    # Set Determined password to something in order to disable auto-login.
    password = get_random_string()
    assert change_user_password(constants.DEFAULT_DETERMINED_USER, password) == 0

    # Log in as determined.
    api_utils.configure_token_store(
        authentication.Credentials(constants.DEFAULT_DETERMINED_USER, password)
    )
    # login test user.
    api_utils.configure_token_store(creds)
    child = det_spawn(["user", "logout", "--all"])
    child.wait()
    child.close()
    assert child.status == 0
    # Trying to list experiments should result in an error.
    child = det_spawn(["e", "list"])
    expected = "Unauthenticated"
    assert expected in str(child.read())
    child.wait()
    child.close()
    assert child.status != 0

    # Log in as determined.
    api_utils.configure_token_store(
        authentication.Credentials(constants.DEFAULT_DETERMINED_USER, password)
    )
    # Change Determined password back to "".
    change_user_password(constants.DEFAULT_DETERMINED_USER, "")


@pytest.mark.e2e_cpu
def test_user_edit(clean_auth: None, login_admin: None) -> None:
    _, original_name, _ = api_utils.create_test_user(False)
    admin = api_utils.admin_session()

    current_user = _fetch_user_by_username(admin, original_name)

    new_display_name = get_random_string()
    new_username = get_random_string()

    assert current_user is not None and current_user.id
    command = [
        "det",
        "user",
        "edit",
        original_name,
        "--display-name",
        new_display_name,
        "--username",
        new_username,
        "--active=true",
        "--remote=false",
        "--admin=true",
    ]
    detproc.check_output(admin, command)

    modded_user = bindings.get_GetUser(admin, userId=current_user.id).user
    assert modded_user is not None
    assert modded_user.displayName == new_display_name
    assert modded_user.username == new_username
    assert modded_user.active
    assert not modded_user.remote
    assert modded_user.admin


@pytest.mark.e2e_cpu
def test_user_edit_no_fields(clean_auth: None, login_admin: None) -> None:
    _, username, _ = api_utils.create_test_user(False)
    admin = api_utils.admin_session()

    # No edited field should result in error
    command = ["det", "user", "edit", username]
    detproc.check_error(admin, command, "No field provided")
