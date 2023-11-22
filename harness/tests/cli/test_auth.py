import contextlib
import json
import os
import shutil
from pathlib import Path
from typing import Any, Iterator, Optional, Tuple
from unittest import mock

import pytest
import requests_mock

from determined.common.api import authentication, certs
from tests.confdir import use_test_config_dir

MOCK_MASTER_URL = "http://localhost:8080"
AUTH_V0_PATH = Path(__file__).parent / "auth_v0.json"
UNTRUSTED_CERT_PATH = Path(__file__).parents[1] / "common" / "untrusted-root" / "127.0.0.1-ca.crt"
AUTH_JSON = {
    "version": 1,
    "masters": {
        "http://localhost:8080": {
            "active_user": "bob",
            "tokens": {
                "determined": "det.token",
                "bob": "bob.token",
            },
        }
    },
}


@pytest.mark.parametrize("user", [None, "bob", "determined"])
def test_auth_with_store(requests_mock: requests_mock.Mocker, user: Optional[str]) -> None:
    with use_test_config_dir() as config_dir:
        auth_json_path = config_dir / "auth.json"
        with open(auth_json_path, "w") as f:
            json.dump(AUTH_JSON, f)

        expected_user = "determined" if user == "determined" else "bob"
        expected_token = "det.token" if user == "determined" else "bob.token"
        requests_mock.get(
            "/api/v1/me",
            status_code=200,
            json={"username": expected_user},
        )
        auth = authentication.Authentication(MOCK_MASTER_URL, user)
        assert auth.session.username == expected_user
        assert auth.session.token == expected_token


class Check:
    """A result in a ScenarioSet indicating a check of a particular token is expected."""

    def __init__(self, token: str) -> None:
        self.token = token

    def __repr__(self) -> str:
        return f"Check({self.token!r})"


class Login:
    """A result in a ScenarioSet indicating a particular login call is expected."""

    def __init__(self, username: str, password: str) -> None:
        self.username = username
        self.password = password

    def __repr__(self) -> str:
        return f"Login({self.username!r}, {self.password!r})"


class Use:
    """
    A result in a ScenarioSet asserting a particular token should be returned.

    This one is only used when the environment token is active, because that code path is designed
    to make no additional calls (why would a container created by the master have to double-check
    that it's token in the environment is valid?).  Without Use, some test would be unclear.
    """

    def __init__(self, token: str) -> None:
        self.token = token

    def __repr__(self) -> str:
        return f"Use({self.token!r})"


def setenv_optional(key: str, val: Optional[str]) -> None:
    if val is None:
        os.environ.pop(key, None)
    else:
        os.environ[key] = val


class ScenarioSet:
    """
    Our login logic is so complex I (rb) can't really keep it all in my head.  So the next best
    thing is to have exhaustive test cases, both to keep the complex system working and to make sure
    changes in the system are possible without such a high risk of accidental breakages.
    """

    def __init__(
        self,
        req_user: str,
        req_pass: str,
        env_user: str,
        env_pass: str,
        env_token: str,
        cache: str,
        *results: Any,
    ) -> None:
        self.req_user = req_user
        self.req_pass = req_pass
        self.env_user = env_user
        self.env_pass = env_pass
        self.env_token = env_token
        self.cache = cache
        self.results = results

    def __repr__(self) -> str:
        results = ", ".join(repr(r) for r in self.results)
        return (
            f"ScenarioSet({self.req_user!r}, {self.req_pass!r}, {self.env_user!r}, "
            f"{self.env_pass!r}, {self.env_token!r}, {self.cache!r}, {results})"
        )

    def scenarios(self) -> Iterator[Tuple[str, str, str, str, str, str]]:
        for req_user in "yn" if self.req_user == "*" else self.req_user:
            for req_pass in "yn" if self.req_pass == "*" else self.req_pass:
                for env_user in "yne" if self.env_user == "*" else self.env_user:
                    for env_pass in "yn" if self.env_pass == "*" else self.env_pass:
                        for env_token in "yn" if self.env_token == "*" else self.env_token:
                            for cache in "ynx" if self.cache == "*" else self.cache:
                                yield req_user, req_pass, env_user, env_pass, env_token, cache


@pytest.mark.parametrize(
    "scenario_set",
    [
        # Explicit user and password, no environment settings, or DET_USER is overridden by
        # explicitly requested user.
        #
        #           requested user (y="user", n=None)
        #            |   requested password (y="req_pass", n=None)
        #            |    |   DET_USER env var (y="user", n=unset, e="extra")
        #            |    |    |    DET_PASS env var (y="env_pass, e=None)
        #            |    |    |     |   DET_USER_TOKEN (y="env_token", e=None)
        #            |    |    |     |    |   token cache state (y="cache", n=empty, x="expired")
        #            |    |    |     |    |    |   Results...
        #            |    |    |     |    |    |    |
        ScenarioSet("y", "y", "ne", "*", "*", "y", Check("cache")),
        ScenarioSet("y", "y", "ne", "*", "*", "n", Login("user", "req_pass")),
        ScenarioSet("y", "y", "ne", "*", "*", "x", Check("expired"), Login("user", "req_pass")),
        # ---
        # Explicit user, but password not provided.  Still no (relevant) environment settings.
        ScenarioSet("y", "n", "ne", "*", "*", "y", Check("cache")),
        ScenarioSet("y", "n", "ne", "*", "*", "n", Login("user", "prompt_pass")),
        ScenarioSet("y", "n", "ne", "*", "*", "x", Check("expired"), Login("user", "prompt_pass")),
        # ---
        # DET_USER_TOKEN is overridden by a configured cache (the on-cluster `det user login` case).
        ScenarioSet("*", "*", "ye", "*", "y", "y", Check("cache")),
        # DET_USER_TOKEN can be used if user is explicit, so long as it matches.
        ScenarioSet("y", "*", "y", "*", "y", "n", Use("env_token")),
        # Token in env is used if cache is missing or invalid.
        ScenarioSet("n", "*", "y", "*", "y", "n", Use("env_token")),
        ScenarioSet("n", "*", "y", "*", "y", "x", Check("expired"), Use("env_token")),
        # Token in env is overridden by a configured cache (the on-cluster `det user login` case).
        ScenarioSet("n", "*", "ye", "*", "y", "y", Check("cache")),
        # ---
        # Explicit user and password, DET_USER set but DET_USER_TOKEN not set.  DET_PASS is ignored.
        ScenarioSet("y", "y", "y", "*", "n", "y", Check("cache")),
        ScenarioSet("y", "y", "y", "*", "n", "n", Login("user", "req_pass")),
        ScenarioSet("y", "y", "y", "*", "n", "x", Check("expired"), Login("user", "req_pass")),
        # ---
        # Explicit user but no password, DET_USER and DET_PASS set, and DET_USER_TOKEN unset.
        # DET_PASS still ignored since DET_USER/DET_PASS are meant to be processed as a unit, and
        # an explicitly-requested username overrides that unit.
        ScenarioSet("y", "n", "y", "y", "n", "y", Check("cache")),
        ScenarioSet("y", "n", "y", "y", "n", "n", Login("user", "prompt_pass")),
        ScenarioSet("y", "n", "y", "y", "n", "x", Check("expired"), Login("user", "prompt_pass")),
        # ---
        # Explicit user but no password, DET_USER set but no other env.  DET_USER is ignored.
        ScenarioSet("y", "n", "y", "n", "n", "y", Check("cache")),
        ScenarioSet("y", "n", "y", "n", "n", "n", Login("user", "prompt_pass")),
        ScenarioSet("y", "n", "y", "n", "n", "x", Check("expired"), Login("user", "prompt_pass")),
        # ---
        # Nothing explicit; DET_USER and DET_PASS are set, DET_USER_TOKEN is unet.
        # Cache continues to work where it matches DET_USER, and there are no password prompts.
        ScenarioSet("n", "*", "y", "y", "n", "y", Check("cache")),
        ScenarioSet("n", "*", "y", "y", "n", "n", Login("user", "env_pass")),
        ScenarioSet("n", "*", "y", "y", "n", "x", Check("expired"), Login("user", "env_pass")),
        # ---
        # Nothing explicit; and DET_USER is ignored without either DET_PASS or DET_USER_TOKEN.
        ScenarioSet("n", "n", "*", "n", "n", "n", Login("determined", "")),
        # the username is taken from the cache but password must be provided again
        ScenarioSet("n", "n", "*", "n", "n", "x", Check("expired"), Login("user", "prompt_pass")),
        ScenarioSet("n", "n", "*", "n", "n", "y", Check("cache")),
        # ---
        # If password is explicit but username is not, we fall back to the default username.
        ScenarioSet("n", "y", "n", "n", "n", "n", Login("determined", "req_pass")),
        # Other pass-but-not-user cases are governed by other effects (see several earlier cases).
        ScenarioSet("n", "y", "*", "*", "*", "y", Check("cache")),
    ],
)
@mock.patch("determined.common.api.authentication._is_token_valid")
@mock.patch("determined.common.api.authentication.do_login")
@mock.patch("getpass.getpass")
@mock.patch("determined.common.api.authentication.TokenStore.get_active_user")
@mock.patch("determined.common.api.authentication.TokenStore.get_token")
@mock.patch("determined.common.api.authentication.TokenStore.drop_user")
@mock.patch("determined.common.api.authentication.TokenStore.set_token")
def test_login_scenarios(
    # Don't care about set_token and it doesn't return anything.
    _: mock.MagicMock,
    # Don't care about drop_user and it doesn't return anything.
    __: mock.MagicMock,
    mock_get_token: mock.MagicMock,
    mock_get_active_user: mock.MagicMock,
    mock_getpass: mock.MagicMock,
    mock_do_login: mock.MagicMock,
    mock_is_token_valid: mock.MagicMock,
    scenario_set: ScenarioSet,
) -> None:
    def getpass(*_: Any) -> str:
        return "prompt_pass"

    def _is_token_valid(master_url: str, token: str, cert: Any) -> bool:
        return token in ["cache", "env_token"]

    def do_login(*_: Any) -> str:
        return "new_token"

    mock_getpass.side_effect = getpass
    mock_is_token_valid.side_effect = _is_token_valid
    mock_do_login.side_effect = do_login

    for ru, rp, eu, ep, et, cache in scenario_set.scenarios():
        # Convert short-codes from case definitions into real values.
        req_user = {"y": "user", "n": None}[ru]
        req_pass = {"y": "req_pass", "n": None}[rp]
        env_user = {"y": "user", "e": "extra", "n": None}[eu]
        env_pass = {"y": "env_pass", "n": None}[ep]
        env_token = {"y": "env_token", "n": None}[et]

        def get_active_user(*_: Any) -> Optional[str]:
            return None if cache == "n" else "user"  # noqa:B023

        def get_token(*_: Any) -> Optional[str]:
            if cache == "n":  # noqa:B023
                return None
            if cache == "x":  # noqa:B023
                return "expired"
            if cache == "y":  # noqa:B023
                return "cache"
            raise ValueError(f"unexpected cache value {cache!r}")  # noqa:B023

        mock_get_active_user.side_effect = get_active_user
        mock_get_token.side_effect = get_token

        mock_is_token_valid.reset_mock()
        mock_do_login.reset_mock()

        old_det_user = os.environ.get("DET_USER")
        old_det_pass = os.environ.get("DET_PASS")
        old_det_token = os.environ.get("DET_USER_TOKEN")
        try:
            setenv_optional("DET_USER", env_user)
            setenv_optional("DET_PASS", env_pass)
            setenv_optional("DET_USER_TOKEN", env_token)

            auth = authentication.Authentication("master_url", req_user, req_pass, None)

            # Make sure we got the results we expected.
            for result in scenario_set.results:
                if isinstance(result, Check):
                    mock_is_token_valid.assert_has_calls(
                        [mock.call("master_url", result.token, None)]
                    )
                elif isinstance(result, Login):
                    mock_do_login.assert_has_calls(
                        [mock.call("master_url", result.username, result.password, None)]
                    )
                elif isinstance(result, Use):
                    assert auth.session.token == result.token
                else:
                    raise ValueError(f"unexpected result: {result}")

            # Make sure we didn't get any unexpected results.
            if not any(isinstance(result, Check) for result in scenario_set.results):
                mock_is_token_valid.assert_not_called()
            if not any(isinstance(result, Login) for result in scenario_set.results):
                mock_do_login.assert_not_called()

        except Exception as e:
            raise RuntimeError(
                f"failed {scenario_set} with ru={ru} rp={rp} eu={eu} ep={ep} et={et} cache={cache}"
            ) from e

        finally:
            setenv_optional("DET_USER", old_det_user)
            setenv_optional("DET_PASS", old_det_pass)
            setenv_optional("DET_USER_TOKEN", old_det_token)


@contextlib.contextmanager
def set_container_env_vars() -> Iterator[None]:
    try:
        os.environ["DET_USER"] = "alice"
        os.environ["DET_USER_TOKEN"] = "alice.token"
        yield
    finally:
        del os.environ["DET_USER"]
        del os.environ["DET_USER_TOKEN"]


@pytest.mark.parametrize("user", [None, "bob", "determined"])
@pytest.mark.parametrize("has_token_store", [True, False])
def test_auth_user_from_env(
    requests_mock: requests_mock.Mocker, user: Optional[str], has_token_store: bool
) -> None:
    with use_test_config_dir() as config_dir, set_container_env_vars():
        if has_token_store:
            auth_json_path = config_dir / "auth.json"
            with open(auth_json_path, "w") as f:
                json.dump(AUTH_JSON, f)

        requests_mock.get("/api/v1/me", status_code=200, json={"username": "alice"})

        if has_token_store:
            nop_password = "user_password"
            auth = authentication.Authentication(MOCK_MASTER_URL, user, nop_password)
            assert auth.session.username == user or "determined"
            assert auth.session.token == ("det.token" if user == "determined" else "bob.token")
        else:
            auth = authentication.Authentication(MOCK_MASTER_URL)
            assert auth.session.username == "alice"
            assert auth.session.token == "alice.token"


def test_auth_json_v0_upgrade() -> None:
    with use_test_config_dir() as config_dir:
        auth_json_path = config_dir / "auth.json"
        shutil.copy2(AUTH_V0_PATH, auth_json_path)
        ts = authentication.TokenStore(MOCK_MASTER_URL, auth_json_path)

        assert ts.get_active_user() == "determined"
        assert ts.get_token("determined") == "v2.public.this.is.a.test"

        ts.set_token("determined", "ai")

        ts2 = authentication.TokenStore(MOCK_MASTER_URL, auth_json_path)
        assert ts2.get_token("determined") == "ai"

        with auth_json_path.open() as fin:
            data = json.load(fin)
            assert data.get("version") == 1
            assert "masters" in data and list(data["masters"].keys()) == [MOCK_MASTER_URL]


def test_cert_v0_upgrade() -> None:
    with use_test_config_dir() as config_dir:
        cert_path = config_dir / "master.crt"
        shutil.copy2(UNTRUSTED_CERT_PATH, cert_path)
        with cert_path.open() as fin:
            cert_data = fin.read()

        cert = certs.default_load(MOCK_MASTER_URL)
        assert isinstance(cert.bundle, str)
        with open(cert.bundle) as fin:
            loaded_cert_data = fin.read()
        assert loaded_cert_data.endswith(cert_data)
        assert not cert_path.exists()

        v1_certs_path = config_dir / "certs.json"
        assert v1_certs_path.exists()

        # Load once again from v1.
        cert2 = certs.default_load(MOCK_MASTER_URL)
        assert isinstance(cert2.bundle, str)
        with open(cert2.bundle) as fin:
            loaded_cert_data = fin.read()
        assert loaded_cert_data.endswith(cert_data)
