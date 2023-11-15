"""
detproc is a subprocess-like tool for calling our CLI with explicit session management.

e2e tests shouldn't really be relying on the persistence of cached api credentials in order to work;
they should be explicit about which login session should be used to make the test pass.

However, lots of e2e functionality is exercised today through the CLI.  Also, it's unfortunately
true that almost all of the CLI functionality is tested only with e2e tests.  So migrating the whole
e2e test suite to api.bindings or the SDK might be nice for e2e_tests but it would probably result
in huge parts of the CLI having no test coverage at all.

So the detprocess module avoides the dilemma by continuing to use the CLI in e2e tests but offering
a mechansism for explicit session management through the CLI subproce process boundary.
"""

import json
import os
import subprocess
from typing import Any, Dict, List, Optional

from determined.common import api


def mkenv(sess: api.Session, env: Optional[Dict[str, str]]) -> Dict[str, str]:
    env = env or {**os.environ}
    assert "DET_USER" not in env, "if you set DET_USER you probably want to use normal subprocess"
    assert (
        "DET_USER_TOKEN" not in env
    ), "if you set DET_USER_TOKEN you probably want to use normal subprocess"
    # Point at the same master as the session.
    env["DET_MASTER"] = sess._master
    # Configure the username and token directly through the environment, via the codepath normally
    # designed for on-cluster auto-config sitautions.
    env["DET_USER"] = sess._utp.username
    env["DET_USER_TOKEN"] = sess._utp.token
    # Disable the authentication cache, which, by design, is allowed to override that on-cluster
    # auto-config situation.
    env["DET_DEBUG_CONFIG_PATH"] = "/tmp/empty"
    return env


class Popen(subprocess.Popen):
    def __init__(
        self,
        sess: api.Session,
        cmd: List[str],
        *args: Any,
        env: Optional[Dict[str, str]] = None,
        **kwargs: Any,
    ) -> None:
        if "-u" in cmd:
            # XXX: is this check a good idea?
            raise ValueError(cmd)
        super().__init__(cmd, *args, env=mkenv(sess, env), **kwargs)


def run(
    sess: api.Session,
    cmd: List[str],
    *args: Any,
    env: Optional[Dict[str, str]] = None,
    **kwargs: Any,
) -> subprocess.CompletedProcess:
    if "-u" in cmd:
        # XXX: is this check a good idea?
        raise ValueError(cmd)
    return subprocess.run(cmd, *args, env=mkenv(sess, env), **kwargs)


def check_call(
    sess: api.Session,
    cmd: List[str],
    env: Optional[Dict[str, str]] = None,
) -> subprocess.CompletedProcess:
    if "-u" in cmd:
        raise ValueError(cmd)
    return subprocess.check_call(cmd, env=mkenv(sess, env))


def check_output(
    sess: api.Session,
    cmd: List[str],
    env: Optional[Dict[str, str]] = None,
) -> str:
    if "-u" in cmd:
        raise ValueError(cmd)
    return subprocess.check_output(cmd, env=mkenv(sess, env)).decode()


def check_error(
    sess: api.Session,
    cmd: List[str],
    errmsg: str,
) -> subprocess.CompletedProcess:
    p = run(sess, cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert p.exit_code != 0
    assert errmsg.lower() in p.stdout.decode("utf8").lower()
    return p


def check_json(
    sess: api.Session,
    cmd: List[str],
) -> Any:
    return json.loads(check_output(sess))
