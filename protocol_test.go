/*
 * Copyright (C) 2015-2018 Virgil Security Inc.
 *
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     (1) Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 *
 *     (2) Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in
 *     the documentation and/or other materials provided with the
 *     distribution.
 *
 *     (3) Neither the name of the copyright holder nor the names of its
 *     contributors may be used to endorse or promote products derived from
 *     this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE AUTHOR ''AS IS'' AND ANY EXPRESS OR
 * IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT,
 * INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
 * STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING
 * IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 *
 * Lead Maintainer: Virgil Security Inc. <support@virgilsecurity.com>
 */

package purekit

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProtocol_EnrollAccount(t *testing.T) {

	req := require.New(t)

	appToken := os.Getenv("APP_TOKEN")

	if appToken == "" {
		t.Skip("no parameters")
	}

	skStr := os.Getenv("SECRET_KEY")

	pubStr := os.Getenv("PUBLIC_KEY")
	token1 := os.Getenv("UPDATE_TOKEN")
	address := os.Getenv("SERVER_ADDRESS")

	context, err := CreateContext(appToken, pubStr, skStr, "")
	req.NoError(err)

	proto, err := NewProtocol(context)
	req.NoError(err)

	if address != "" {
		proto.APIClient = &APIClient{
			AppToken: appToken,
			URL:      address,
		}
	}

	const pwd = "p@ssw0Rd"
	//enroll version 1
	rec, key, err := proto.EnrollAccount(pwd)
	req.NoError(err)
	req.True(len(rec) > 0)
	req.True(len(key) == 32)
	//verify version 1
	key1, err := proto.VerifyPassword(pwd, rec)
	req.NoError(err)
	req.Equal(key, key1)

	key2, err := proto.VerifyPassword("p@ss", rec)
	req.EqualError(err, ErrInvalidPassword.Error())
	req.Nil(key2)

	//rotate happened
	context, err = CreateContext(appToken, pubStr, skStr, token1)
	req.NoError(err)
	proto, err = NewProtocol(context)
	req.NoError(err)

	if address != "" {
		proto.APIClient = &APIClient{
			AppToken: appToken,
			URL:      address,
		}
	}

	time.Sleep(2 * time.Second)
	//verify version 1 with token
	key3, err := proto.VerifyPassword(pwd, rec)
	req.NoError(err)
	req.Equal(key, key3)

	updater, err := NewRecordUpdater(token1)
	req.NoError(err)
	newRec, err := updater.UpdateRecord(rec)
	req.NoError(err)
	//verify version 2
	key4, err := proto.VerifyPassword(pwd, newRec)
	req.NoError(err)
	req.Equal(key, key4)

	//enroll version 2
	rec, key, err = proto.EnrollAccount("passw0rd")
	req.NoError(err)

	version, _, err := UnmarshalRecord(rec)
	req.NoError(err)
	req.Equal(version, uint32(2))

	//verify version 2
	key2, err = proto.VerifyPassword("passw0rd", rec)
	req.NoError(err)
	req.Equal(key2, key)
}
