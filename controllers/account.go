// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	ResponseTypeLogin = "login"
	ResponseTypeCode  = "code"
)

type RequestForm struct {
	Type string `json:"type"`

	Organization string `json:"organization"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`

	Application string `json:"application"`
	Provider    string `json:"provider"`
	Code        string `json:"code"`
	State       string `json:"state"`
	RedirectUri string `json:"redirectUri"`
	Method      string `json:"method"`
}

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

// @Title Signup
// @Description sign up a new user
// @Param   username     formData    string  true        "The username to sign up"
// @Param   password     formData    string  true        "The password"
// @Success 200 {object} controllers.Response The Response object
// @router /signup [post]
func (c *ApiController) Signup() {
	var resp Response

	if c.GetSessionUser() != "" {
		resp = Response{Status: "error", Msg: "Please log out first before signing up", Data: c.GetSessionUser()}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		panic(err)
	}

	msg := object.CheckUserSignup(form.Username, form.Password)
	if msg != "" {
		resp = Response{Status: "error", Msg: msg, Data: ""}
	} else {
		user := &object.User{
			Owner:        form.Organization,
			Name:         form.Username,
			CreatedTime:  util.GetCurrentTime(),
			Password:     form.Password,
			PasswordType: "plain",
			DisplayName:  form.Name,
			Email:        form.Email,
			Phone:        form.Phone,
		}
		object.AddUser(user)

		//c.SetSessionUser(user)

		util.LogInfo(c.Ctx, "API: [%s] is signed up as new user", user)
		resp = Response{Status: "ok", Msg: "", Data: user}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Logout
// @Description logout the current user
// @Success 200 {object} controllers.Response The Response object
// @router /logout [post]
func (c *ApiController) Logout() {
	var resp Response

	user := c.GetSessionUser()
	util.LogInfo(c.Ctx, "API: [%s] logged out", user)

	c.SetSessionUser("")

	resp = Response{Status: "ok", Msg: "", Data: user}

	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title GetAccount
// @Description get the details of the current account
// @Success 200 {object} controllers.Response The Response object
// @router /get-account [get]
func (c *ApiController) GetAccount() {
	var resp Response

	if c.GetSessionUser() == "" {
		resp = Response{Status: "error", Msg: "Please sign in first", Data: c.GetSessionUser()}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	username := c.GetSessionUser()
	user := object.GetUser(username)
	resp = Response{Status: "ok", Msg: "", Data: user}

	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title UploadAvatar
// @Description upload avatar
// @Param   avatarfile   formData    string  true        "The base64 encode of avatarfile"
// @Param   password     formData    string  true        "The password"
// @Success 200 {object} controllers.Response The Response object
// @router /upload-avatar [post]
func (c *ApiController) UploadAvatar() {
	var resp Response

	username := c.GetSessionUser()
	if c.GetSessionUser() == "" {
		resp = Response{Status: "error", Msg: "Please sign in first", Data: c.GetSessionUser()}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	user := object.GetUser(username)

	avatarBase64 := c.Ctx.Request.Form.Get("avatarfile")
	index := strings.Index(avatarBase64, ",")
	if index < 0 || avatarBase64[0:index] != "data:image/png;base64" {
		resp = Response{Status: "error", Msg: "File encoding error"}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	dist, _ := base64.StdEncoding.DecodeString(avatarBase64[index+1:])
	msg := object.UploadAvatar(user.Name, dist)
	if msg != "" {
		resp = Response{Status: "error", Msg: msg}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}
	user.Avatar = fmt.Sprintf("%s%s.png?time=%s", object.GetAvatarPath(), user.Name, util.GetCurrentUnixTime())
	object.UpdateUser(user.Owner+"/"+user.Name, user)
	resp = Response{Status: "ok", Msg: ""}
	c.Data["json"] = resp
	c.ServeJSON()
}
