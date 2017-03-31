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
	"fmt"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mendersoftware/go-lib-micro/accesslog"
	mcontext "github.com/mendersoftware/go-lib-micro/context"
	"github.com/mendersoftware/go-lib-micro/customheader"
	dlog "github.com/mendersoftware/go-lib-micro/log"
	"github.com/mendersoftware/go-lib-micro/requestid"
	"github.com/mendersoftware/go-lib-micro/requestlog"

	api_http "github.com/mendersoftware/useradm/api/http"
	"github.com/mendersoftware/useradm/authz"
	"github.com/mendersoftware/useradm/jwt"
)

const (
	EnvProd = "prod"
	EnvDev  = "dev"
)

var (
	DefaultDevStack = []rest.Middleware{

		// logging
		&requestlog.RequestLogMiddleware{},
		&accesslog.AccessLogMiddleware{Format: accesslog.SimpleLogFormat},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},

		// catches the panic errors that occur with stack trace
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: true,
		},

		// json pretty print
		&rest.JsonIndentMiddleware{},

		// verifies the request Content-Type header
		// The expected Content-Type is 'application/json'
		// if the content is non-null
		&rest.ContentTypeCheckerMiddleware{},
		&requestid.RequestIdMiddleware{},
	}

	DefaultProdStack = []rest.Middleware{

		// logging
		&requestlog.RequestLogMiddleware{},
		&accesslog.AccessLogMiddleware{Format: accesslog.SimpleLogFormat},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},

		// catches the panic errors
		&rest.RecoverMiddleware{},

		// response compression
		&rest.GzipMiddleware{},

		// verifies the request Content-Type header
		// The expected Content-Type is 'application/json'
		// if the content is non-null
		&rest.ContentTypeCheckerMiddleware{},
		&requestid.RequestIdMiddleware{},
	}

	middlewareMap = map[string][]rest.Middleware{
		EnvProd: DefaultProdStack,
		EnvDev:  DefaultDevStack,
	}
)

func SetupMiddleware(api *rest.Api, mwtype string, authorizer authz.Authorizer, jwth jwt.JWTHandler) error {

	l := dlog.New(dlog.Ctx{})

	api.Use(&customheader.CustomHeaderMiddleware{
		HeaderName:  "X-USERADM-VERSION",
		HeaderValue: CreateVersionString(),
	})

	l.Infof("setting up %s middleware", mwtype)

	mwstack, ok := middlewareMap[mwtype]
	if !ok {
		return fmt.Errorf("incorrect middleware type: %s", mwtype)
	}

	api.Use(mwstack...)

	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,

		// Should be tested with some list
		OriginValidator: func(origin string, request *rest.Request) bool {
			// Accept all requests
			return true
		},

		// Preflight request cache length
		AccessControlMaxAge: 60,

		// Allow authentication requests
		AccessControlAllowCredentials: true,

		// Allowed headers
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},

		// Allowed headers
		AllowedHeaders: []string{
			"Accept",
			"Allow",
			"Content-Type",
			"Origin",
			"Authorization",
			"Accept-Encoding",
			"Access-Control-Request-Headers",
			"Header-Access-Control-Request",
		},

		// Headers that can be exposed to JS
		AccessControlExposeHeaders: []string{
			"Location",
			"Link",
		},
	})

	// populate request context with data from rest.Request.Env just before
	// going to AuthzMiddleware/IfMiddleware
	api.Use(&mcontext.UpdateContextMiddleware{
		Updates: []mcontext.UpdateContextFunc{
			mcontext.RepackLoggerToContext,
			mcontext.RepackRequestIdToContext,
		},
	})

	authzmw := &authz.AuthzMiddleware{
		Authz:      authorizer,
		ResFunc:    api_http.ExtractResourceAction,
		JWTHandler: jwth,
	}

	//force authz only on verification endpoint
	ifmw := &rest.IfMiddleware{
		Condition: api_http.IsVerificationEndpoint,
		IfTrue:    authzmw,
	}

	api.Use(ifmw)

	return nil
}
