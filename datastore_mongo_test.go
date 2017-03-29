// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"context"
	"testing"

	"github.com/mendersoftware/go-lib-micro/mongo/migrate"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestMongoIsEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	testCases := map[string]struct {
		empty bool
	}{
		"empty": {
			empty: true,
		},
		"not empty": {
			empty: false,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		session := db.Session()
		store, err := NewDataStoreMongoWithSession(session)
		assert.NoError(t, err)

		if !tc.empty {
			// insert anything
			session.DB(DbName).C(DbUsersColl).Insert(tc)
		}

		empty, err := store.IsEmpty()

		assert.Equal(t, tc.empty, empty)
		assert.NoError(t, err)

		// Need to close all sessions to be able to call wipe at next
		// test case
		session.Close()
	}
}

func TestMongoCreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	exisitingUsers := []interface{}{
		UserModel{
			ID:       "1",
			Email:    "foo@bar.com",
			Password: "pretenditsahash",
		},
		UserModel{
			ID:       "2",
			Email:    "bar@bar.com",
			Password: "pretenditsahash",
		},
	}

	testCases := map[string]struct {
		inUser UserModel
		outErr string
	}{
		"ok": {
			inUser: UserModel{
				ID:       "1234",
				Email:    "baz@bar.com",
				Password: "correcthorsebatterystaple",
			},
			outErr: "",
		},
		"duplicate email error": {
			inUser: UserModel{
				ID:       "1234",
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			outErr: "user with a given email already exists",
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		db.Wipe()

		session := db.Session()
		store, err := NewDataStoreMongoWithSession(session)
		assert.NoError(t, err)

		err = session.DB(DbName).C(DbUsersColl).Insert(exisitingUsers...)
		assert.NoError(t, err)

		pass := tc.inUser.Password
		err = store.CreateUser(&tc.inUser)

		if tc.outErr == "" {
			//fetch user by id, verify password checks out
			var user UserModel
			err := session.DB(DbName).C(DbUsersColl).FindId("1234").One(&user)
			assert.NoError(t, err)
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))

			assert.NoError(t, err)

		} else {
			assert.EqualError(t, err, tc.outErr)
		}

		session.Close()
	}
}

func TestMongoGetUserByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	existingUsers := []interface{}{
		UserModel{
			ID:       "1",
			Email:    "foo@bar.com",
			Password: "passwordhash12345",
		},
		UserModel{
			ID:       "2",
			Email:    "bar@bar.com",
			Password: "passwordhashqwerty",
		},
	}

	testCases := map[string]struct {
		inEmail string
		outUser *UserModel
	}{
		"ok - found 1": {
			inEmail: "foo@bar.com",
			outUser: &UserModel{
				ID:       "1",
				Email:    "foo@bar.com",
				Password: "passwordhash12345",
			},
		},
		"ok - found 2": {
			inEmail: "bar@bar.com",
			outUser: &UserModel{
				ID:       "2",
				Email:    "bar@bar.com",
				Password: "passwordhashqwerty",
			},
		},
		"not found": {
			inEmail: "baz@bar.com",
			outUser: nil,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		db.Wipe()

		session := db.Session()
		store, err := NewDataStoreMongoWithSession(session)
		assert.NoError(t, err)

		err = session.DB(DbName).C(DbUsersColl).Insert(existingUsers...)
		assert.NoError(t, err)

		user, err := store.GetUserByEmail(tc.inEmail)

		if tc.outUser != nil {
			assert.Equal(t, *tc.outUser, *user)
		} else {
			assert.Nil(t, user)
			assert.Nil(t, err)
		}

		session.Close()
	}
}

func TestMongoGetUserById(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	existingUsers := []interface{}{
		UserModel{
			ID:       "1",
			Email:    "foo@bar.com",
			Password: "passwordhash12345",
		},
		UserModel{
			ID:       "2",
			Email:    "bar@bar.com",
			Password: "passwordhashqwerty",
		},
	}

	testCases := map[string]struct {
		inId    string
		outUser *UserModel
	}{
		"ok - found 1": {
			inId: "1",
			outUser: &UserModel{
				ID:       "1",
				Email:    "foo@bar.com",
				Password: "passwordhash12345",
			},
		},
		"ok - found 2": {
			inId: "2",
			outUser: &UserModel{
				ID:       "2",
				Email:    "bar@bar.com",
				Password: "passwordhashqwerty",
			},
		},
		"not found": {
			inId:    "3",
			outUser: nil,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		db.Wipe()

		session := db.Session()
		store, err := NewDataStoreMongoWithSession(session)
		assert.NoError(t, err)

		err = session.DB(DbName).C(DbUsersColl).Insert(existingUsers...)
		assert.NoError(t, err)

		user, err := store.GetUserById(tc.inId)

		if tc.outUser != nil {
			assert.Equal(t, *tc.outUser, *user)
		} else {
			assert.Nil(t, user)
			assert.Nil(t, err)
		}

		session.Close()
	}
}

func TestMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigrate in short mode.")
	}

	testCases := map[string]struct {
		version string
		err     string
	}{
		"0.1.0": {
			version: "0.1.0",
			err:     "",
		},
		"1.2.3": {
			version: "1.2.3",
			err:     "",
		},
		"0.1 error": {
			version: "0.1",
			err:     "failed to parse service version: failed to parse Version: unexpected EOF",
		},
	}

	for name, tc := range testCases {
		t.Logf("case: %s", name)
		db.Wipe()
		session := db.Session()

		store, err := NewDataStoreMongoWithSession(session)
		assert.NoError(t, err)

		ctx := context.Background()

		err = store.Migrate(ctx, tc.version, nil)
		if tc.err == "" {
			assert.NoError(t, err)
			var out []migrate.MigrationEntry
			session.DB(DbName).C(migrate.DbMigrationsColl).Find(nil).All(&out)
			assert.Len(t, out, 1)
			v, _ := migrate.NewVersion(tc.version)
			assert.Equal(t, *v, out[0].Version)
		} else {
			assert.EqualError(t, err, tc.err)
		}

		session.Close()
	}

}
