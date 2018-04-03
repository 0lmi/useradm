// Copyright 2018 Northern.tech AS
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
package jwt

import (
	"time"
)

type Claims struct {
	Audience  string `json:"aud,omitempty" bson:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty" bson:"exp,omitempty"`
	ID        string `json:"jti,omitempty" bson:"jti,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty" bson:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty" bson:"iss,omitempty"`
	NotBefore int64  `json:"nbf,omitempty" bson:"nbf,omitempty"`
	Subject   string `json:"sub,omitempty" bson:"sub,omitempty"`
	Scope     string `json:"scp,omitempty" bson:"scp,omitempty"`
	Tenant    string `json:"mender.tenant,omitempty" bson:"tenant,omitempty"`
	User      bool   `json:"mender.user,omitempty" bson:"user,omitempty"`
}

// Valid checks if claims are valid. Returns error if validation fails.
// Note that for now we're only using iss, exp, sub, scp.
// Basic checks are done here, field correctness (e.g. issuer) - at the service level, where this info is available.
func (c *Claims) Valid() error {
	if c.Issuer == "" ||
		c.ExpiresAt == 0 ||
		c.Subject == "" ||
		c.Scope == "" {
		return ErrTokenInvalid
	}

	if !verifyExp(c.ExpiresAt) {
		return ErrTokenExpired
	}

	return nil
}

func verifyExp(exp int64) bool {
	now := time.Now().Unix()
	return now <= exp
}
