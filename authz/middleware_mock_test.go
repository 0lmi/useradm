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
package authz

import "github.com/stretchr/testify/mock"

import "github.com/mendersoftware/go-lib-micro/log"

// Authorizer is an autogenerated mock type for the Authorizer type
type MockAuthorizer struct {
	mock.Mock
}

// Authorize provides a mock function with given fields: token, resource, action
func (_m *MockAuthorizer) Authorize(token string, resource string, action string) error {
	ret := _m.Called(token, resource, action)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(token, resource, action)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WithLog provides a mock function with given fields: l
func (_m *MockAuthorizer) WithLog(l *log.Logger) Authorizer {
	ret := _m.Called(l)

	var r0 Authorizer
	if rf, ok := ret.Get(0).(func(*log.Logger) Authorizer); ok {
		r0 = rf(l)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Authorizer)
		}
	}

	return r0
}
