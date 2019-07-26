/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package service

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestParsePayload(t *testing.T) {
	code, _ := hex.DecodeString("00c66b6a143b35f2816fe9176e6f3b954e94ca198cd0e9bdffc86a4230336163656137353865343962383761303362336462663239653330353538353763653761343637336561383634653634306564386631336434333836316461343151c1c86a02204e51c1c86c12756e417574686f72697a65466f72506565721400000000000000000000000000000000000000070068164f6e746f6c6f67792e4e61746976652e496e766f6b65")
	_, err := ParsePayload(code)
	fmt.Println("##", err)
}
