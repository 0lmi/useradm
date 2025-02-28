// Copyright 2022 Northern.tech AS
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
	"crypto/rsa"

	jwtgo "github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

var (
	ErrTokenExpired = errors.New("jwt: token expired")
	ErrTokenInvalid = errors.New("jwt: token invalid")
)

// JWTHandler jwt generator/verifier
//go:generate ../utils/mockgen.sh
type Handler interface {
	ToJWT(t *Token) (string, error)
	// FromJWT parses the token and does basic validity checks (Claims.Valid().
	// returns:
	// ErrTokenExpired when the token is valid but expired
	// ErrTokenInvalid when the token is invalid (malformed, missing required claims, etc.)
	FromJWT(string) (*Token, error)
}

// JWTHandlerRS256 is an RS256-specific JWTHandler
type JWTHandlerRS256 struct {
	privKey         *rsa.PrivateKey
	fallbackPrivKey *rsa.PrivateKey
}

func NewJWTHandlerRS256(privKey *rsa.PrivateKey, fallbackPrivKey *rsa.PrivateKey) *JWTHandlerRS256 {
	return &JWTHandlerRS256{
		privKey:         privKey,
		fallbackPrivKey: fallbackPrivKey,
	}
}

func (j *JWTHandlerRS256) ToJWT(token *Token) (string, error) {
	//generate
	jt := jwtgo.NewWithClaims(jwtgo.SigningMethodRS256, &token.Claims)

	//sign
	data, err := jt.SignedString(j.privKey)
	return data, err
}

func (j *JWTHandlerRS256) FromJWT(tokstr string) (*Token, error) {
	var err error
	var jwttoken *jwtgo.Token
	for _, privKey := range []*rsa.PrivateKey{
		j.privKey,
		j.fallbackPrivKey,
	} {
		if privKey != nil {
			jwttoken, err = jwtgo.ParseWithClaims(tokstr, &Claims{},
				func(token *jwtgo.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwtgo.SigningMethodRSA); !ok {
						return nil, errors.New("unexpected signing method: " + token.Method.Alg())
					}
					return &privKey.PublicKey, nil
				},
			)
			if jwttoken != nil && err == nil {
				break
			}
		}
	}

	// our Claims return Mender-specific validation errors
	// go-jwt will wrap them in a generic ValidationError - unwrap and return directly
	if err != nil {
		err, ok := err.(*jwtgo.ValidationError)
		if ok && err.Inner != nil {
			return nil, err.Inner
		}
		return nil, err
	}

	token := Token{}

	if claims, ok := jwttoken.Claims.(*Claims); ok && jwttoken.Valid {
		token.Claims = *claims
		return &token, nil
	}
	return nil, ErrTokenInvalid
}
