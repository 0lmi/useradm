#!/usr/bin/python
# Copyright 2016 Mender Software AS
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#
#        https://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.
from common import init_users, init_users_f, init_users_mt, cli,api_client_mgmt, mongo, make_auth
import bravado
import pytest
import tenantadm

class TestManagementApiPostUsersBase:
    def _do_test_ok(self, api_client_mgmt, init_users, new_user, tenant_id=None):
        auth=None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)


        _, r = api_client_mgmt.create_user(new_user, auth)
        assert r.status_code == 201

        users = api_client_mgmt.get_users(auth)
        assert len(users) == len(init_users) + 1

        found_user = [u for u in users if u.email == new_user["email"]]
        assert len(found_user) == 1
        found_user = found_user[0]

    def _do_test_fail_duplicate_email(self, api_client_mgmt, init_users, new_user, tenant_id=None):
        auth=None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        try:
            api_client_mgmt.create_user(new_user, auth)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 422


class TestManagementApiPostUsers(TestManagementApiPostUsersBase):
    def test_ok(self, api_client_mgmt, init_users):
        new_user = {"email":"foo@bar.com", "password": "asdf1234zxcv"}
        self._do_test_ok(api_client_mgmt, init_users, new_user)

    def test_fail_malformed_body(self, api_client_mgmt):
        new_user = {"foo":"bar"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_no_password(self, api_client_mgmt):
        new_user = {"email":"foobar"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_no_email(self, api_client_mgmt):
        new_user = {"password": "asdf1234zxcv"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_not_an_email(self, api_client_mgmt):
        new_user = {"email":"foobar", "password": "asdf1234zxcv"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_pwd_too_short(self, api_client_mgmt):
        new_user = {"email":"foo@bar.com", "password": "asdf"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 422

    def test_fail_duplicate_email(self, api_client_mgmt, init_users):
        new_user = {"email":"foo@bar.com", "password": "asdf"}
        self._do_test_fail_duplicate_email(api_client_mgmt, init_users, new_user)


class TestManagementApiPostUsersMultitenant(TestManagementApiPostUsersBase):
    @pytest.mark.parametrize("tenant_id", ["tenant1id", "tenant2id"])
    def test_ok(self, tenant_id, api_client_mgmt, init_users_mt):
        new_user = {"email":"foo@bar.com", "password": "asdf1234zxcv"}
        with tenantadm.run_fake_create_user(new_user):
            self._do_test_ok(api_client_mgmt, init_users_mt[tenant_id], new_user, tenant_id)

    @pytest.mark.parametrize("tenant_id", ["tenant1id", "tenant2id"])
    def test_fail_duplicate_email(self, tenant_id, api_client_mgmt, init_users_mt):
        new_user = {"email":"foo@bar.com", "password": "asdf1234zxcv"}
        with tenantadm.run_fake_create_user(new_user, 422):
            self._do_test_fail_duplicate_email(api_client_mgmt, init_users_mt[tenant_id], new_user, tenant_id)

class TestManagementApiGetUserBase:
    def _do_test_ok(self, api_client_mgmt, init_users, tenant_id=None):
        auth=None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        for u in init_users:
            found = api_client_mgmt.get_user(u.id, auth)
            assert found.id == u.id
            assert found.email == u.email
            assert found.created_ts == u.created_ts
            assert found.updated_ts == u.updated_ts

    def _do_test_fail_not_found(self, api_client_mgmt, init_users, tenant_id=None):
        auth=None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        try:
            not_found = api_client_mgmt.get_user("madeupid", auth)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 404


class TestManagementApiGetUser(TestManagementApiGetUserBase):
    def test_ok(self, api_client_mgmt, init_users):
        self._do_test_ok(api_client_mgmt, init_users)

    def test_fail_not_found(self, api_client_mgmt, init_users):
        self._do_test_fail_not_found(api_client_mgmt, init_users)


class TestManagementApiGetUserMultitenant(TestManagementApiGetUserBase):
    @pytest.mark.parametrize("tenant_id", ["tenant1id", "tenant2id"])
    def test_ok(self, tenant_id, api_client_mgmt, init_users_mt):
        self._do_test_ok(api_client_mgmt, init_users_mt[tenant_id], tenant_id)

    @pytest.mark.parametrize("tenant_id", ["tenant1id", "tenant2id"])
    def test_fail_not_found(self, tenant_id, api_client_mgmt, init_users_mt):
        self._do_test_fail_not_found(api_client_mgmt, init_users_mt[tenant_id], tenant_id)


class TestManagementApiGetUsersBase:
    def _do_test_ok(self, api_client_mgmt, init_users, tenant_id=None):
        auth=None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        users = api_client_mgmt.get_users(auth)
        assert len(users) == len(init_users)

    def _do_test_no_users(self, api_client_mgmt, tenant_id=None):
        auth=None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        users = api_client_mgmt.get_users(auth)
        assert len(users) == 0

class TestManagementApiGetUsersOk(TestManagementApiGetUsersBase):
    def test_ok(self, api_client_mgmt, init_users):
        self._do_test_ok(api_client_mgmt, init_users)

class TestManagementApiGetUsersNoUsers(TestManagementApiGetUsersBase):
    def test_no_users(self, api_client_mgmt):
        self._do_test_no_users(api_client_mgmt)

class TestManagementApiGetUsersMultitenant(TestManagementApiGetUsersBase):
    @pytest.mark.parametrize("tenant_id", ["tenant1id", "tenant2id"])
    def test_ok(self, tenant_id, api_client_mgmt, init_users_mt):
        self._do_test_ok(api_client_mgmt, init_users_mt[tenant_id], tenant_id)

    @pytest.mark.parametrize("tenant_id", ["tenant1id", "tenant2id"])
    def test_no_users(self, tenant_id, api_client_mgmt, init_users_mt):
        self._do_test_no_users(api_client_mgmt, "non_existing_tenant_id")


class TestManagementApiDeleteUserBase:
    def _do_test_ok(self, api_client_mgmt, init_users, tenant_id=None):
        auth=None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        rsp = api_client_mgmt.delete_user(init_users[0]['id'], auth)
        assert rsp.status_code == 204

        users = api_client_mgmt.get_users(auth)
        assert len(users) == len(init_users) - 1

        found = [u for u in users if u.id == init_users[0]['id']]
        assert len(found) == 0

    def _do_test_not_found(self, api_client_mgmt, tenant_id=None):
        auth=None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        rsp = api_client_mgmt.delete_user('nonexistent_id', auth)
        assert rsp.status_code == 204


class TestManagementApiDeleteUser(TestManagementApiDeleteUserBase):
    def test_ok(self, api_client_mgmt, init_users):
        self._do_test_ok(api_client_mgmt, init_users)

    def test_not_found(self, api_client_mgmt, init_users):
        self._do_test_not_found(api_client_mgmt)


class TestManagementApiDeleteUserMultitenant(TestManagementApiDeleteUserBase):
    @pytest.mark.parametrize("tenant_id", ["tenant1id", "tenant2id"])
    def test_ok(self, tenant_id, api_client_mgmt, init_users_mt):
        with tenantadm.run_fake_delete_user(tenant_id, init_users_mt[tenant_id][0]['id']):
            self._do_test_ok(api_client_mgmt, init_users_mt[tenant_id], tenant_id)

    @pytest.mark.parametrize("tenant_id", ["tenant1id", "tenant2id"])
    def test_not_found(self, tenant_id, api_client_mgmt):
        with tenantadm.run_fake_delete_user():
            self._do_test_not_found(api_client_mgmt, tenant_id)


class TestManagementApiPutUser:
    def test_ok_email(self, api_client_mgmt, init_users_f):
        # test update
        email = "unique1@foo.com"
        update = {"email": email}
        _, r = api_client_mgmt.update_user(init_users_f[0].id, update)
        assert r.status_code == 204

        # get/verify users
        users = api_client_mgmt.get_users()
        assert len(users) == len(init_users_f)

        found = [u for u in users if u.email == email]
        assert len(found) == 1

    def test_ok_pass(self, api_client_mgmt, init_users_f):
        password = "secretpassword123"

        # test update
        update = {"password": password}
        _, r = api_client_mgmt.update_user(init_users_f[1].id, update)
        assert r.status_code == 204

        # get/verify users
        users = api_client_mgmt.get_users()
        assert len(users) == len(init_users_f)

        found = [u for u in users if u.email == init_users_f[1].email]
        assert len(found) == 1

        # try if login still works
        _, r = api_client_mgmt.login(init_users_f[1].email, password)
        assert r.status_code == 200

    def test_ok_email_and_pass(self, api_client_mgmt, init_users_f):
        email = "definitelyunique@foo.com"
        password = "secretpassword123"

        # test update
        update = {"email": email, "password": password}
        _, r = api_client_mgmt.update_user(init_users_f[2].id, update)
        assert r.status_code == 204

        # get/verify users
        users = api_client_mgmt.get_users()
        assert len(users) == len(init_users_f)

        found = [u for u in users if u.email == email]
        assert len(found) == 1

        # try if login still works
        _, r = api_client_mgmt.login(email, password)
        assert r.status_code == 200

    def test_fail_not_found(self, api_client_mgmt, init_users_f):
        update = {"email": "foo@bar.com", "password": "secretpassword123"}

        try:
            _, r = api_client_mgmt.update_user("notfound", update)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 404

    def test_fail_bad_update(self, api_client_mgmt, init_users_f):
        update = {"foo": "bar"}

        try:
            _, r = api_client_mgmt.update_user(init_users_f[3].id, update)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_duplicate_email(self, api_client_mgmt, init_users_f):
        update = {"email": init_users_f[3].email, "password": "secretpassword123"}

        try:
            _, r = api_client_mgmt.update_user(init_users_f[4].id, update)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 422
