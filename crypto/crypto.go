// Copyright (C) 2018 go-gamc authors
//
// This file is part of the go-gamc library.
//
// the go-gamc library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-gamc library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-gamc library.  If not, see <http://www.gnu.org/licenses/>.
//
package crypto

import (
	"gamc.pro/gamcio/go-gamc/crypto/ed25519"
	"gamc.pro/gamcio/go-gamc/crypto/keystore"
	"errors"
)

var (
	// ErrAlgorithmInvalid invalid Algorithm for sign.
	ErrAlgorithmInvalid = errors.New("invalid Algorithm")
)

// NewPrivateKey generate a privatekey
func NewPrivateKey(data []byte) (keystore.PrivateKey, error) {
	var (
		priv *ed25519.PrivateKey
		err  error
	)
	if len(data) == 0 {
		priv = ed25519.NewPrivateKey()
	} else {
		priv = new(ed25519.PrivateKey)
		err = priv.Decode(data)
	}
	if err != nil {
		return nil, err
	}
	return priv, nil
}

func NewPrivateKeyFromSeed(seed []byte) (keystore.PrivateKey, error) {
	return ed25519.NewPrivateKeyFromSeed(seed)
}

// NewSignature returns a ed25519 signature
func NewSignature() (keystore.Signature, error) {
	return new(ed25519.Signature), nil
}
