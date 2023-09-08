import contextlib
import json
import os
import shutil
import unittest.mock
from pathlib import Path
from typing import Iterator, Optional

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
@unittest.mock.patch("determined.common.api.authentication._is_token_valid")
def test_login_with_cache(
    is_token_valid_mock: unittest.mock.MagicMock, user: Optional[str]
) -> None:
    is_token_valid_mock.return_value = True
    with use_test_config_dir() as config_dir:
        auth_json_path = config_dir / "auth.json"
        with open(auth_json_path, "w") as f:
            json.dump(AUTH_JSON, f)

        expected_user = "determined" if user == "determined" else "bob"
        expected_token = "det.token" if user == "determined" else "bob.token"
        utp = authentication.login_with_cache(MOCK_MASTER_URL, user)
        assert utp.username == expected_user
        assert utp.token == expected_token


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
def test_login_user_from_env(
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
            utp = authentication.login_with_cache(MOCK_MASTER_URL, user, nop_password)
            assert utp.username == user or "determined"
            assert utp.token == (
                "det.token" if user == "determined" else "bob.token"
            )
        else:
            utp = authentication.login_with_cache(MOCK_MASTER_URL)
            assert utp.username == "alice"
            assert utp.token == "alice.token"


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
