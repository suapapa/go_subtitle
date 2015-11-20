// Copyright 2013-2015, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtitle

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

type smiTest struct {
	raw string
	exp Book
}

var (
	mil = time.Millisecond

	t1 = smiTest{
		raw: `<SAMI>
<HEAD>
<TITLE>Tae wan Kim</TITLE>
<STYLE TYPE="text/css">
<!--
P { margin-left:8pt; margin-right:8pt; margin-bottom:2pt;
    margin-top:2pt; font-size:14pt; text-align:center;
    font-family:굴림, Arial; font-weight:normal; color:white;
    background-color:black; }
.KRCC { Name:KRCC; lang:kr-KR; SAMIType:CC; }
-->`,
		exp: Book{
			Script{0, 0, 0,
				"Tae wan Kim",
			},
		},
	}

	t2 = smiTest{
		raw: `<SYNC Start=28500><P Class=KRCC>
<font color=66CCFF>이웃집 토토로</font>
<SYNC Start=28997><P Class=KRCC>
<font color=66CCFF>이웃집 토토로</font><br>
<font color=CCFF66>걷자 걷자 난 건강해</font>
<SYNC Start=36741><P Class=KRCC>
<font color=CCFF66>걷는 건 정말 좋아
`,
		exp: Book{
			Script{0,
				28500 * mil,
				28997 * mil,
				"이웃집 토토로",
			},
			Script{0,
				28997 * mil,
				36741 * mil,
				"이웃집 토토로\n걷자 걷자 난 건강해",
			},
			Script{0,
				36741 * mil,
				0,
				"걷는 건 정말 좋아",
			},
		},
	}

	t3 = smiTest{
		raw: `<SYNC Start=18><P Class=KRCC>
고맙다
<SYNC Start=20><P Class=KRCC>&nbsp;
<SYNC Start=24><P Class=KRCC>
자, 다 왔다
<SYNC Start=28><P Class=KRCC>&nbsp;
<SYNC Start=29><P Class=KRCC>
기다려!`,
		exp: Book{
			Script{0,
				18 * mil,
				20 * mil,
				"고맙다",
			},
			Script{0,
				24 * mil,
				28 * mil,
				"자, 다 왔다",
			},
			Script{0,
				29 * mil,
				0 * mil,
				"기다려!",
			},
		},
	}

	t4 = smiTest{
		raw: `<SYNC Start=808211><P Class=KRCC>
오늘은 모두들 논에서 바빴다우<br>
하지만 조금씩 치웠지요
<SYNC Start=817269><P Class=KRCC>
검뎅이 귀신이 도망쳤어!`,
		exp: Book{
			Script{0,
				808211 * mil,
				817269 * mil,
				"오늘은 모두들 논에서 바빴다우\n하지만 조금씩 치웠지요",
			},
			Script{0,
				817269 * mil,
				0,
				"검뎅이 귀신이 도망쳤어!",
			},
		},
	}
)

func TestReadSmiComment(t *testing.T) {
	b, err := ReadSmi(bytes.NewReader([]byte(t1.raw)))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b, t1.exp) {
		t.Log("got:", b)
		t.Log("exp:", t1.exp)
		t.Error("Read comment failed")
	}
}

func TestReadSmiBR(t *testing.T) {
	b, err := ReadSmi(bytes.NewReader([]byte(t2.raw)))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b, t2.exp) {
		t.Log("got:", b)
		t.Log("exp:", t2.exp)
		t.Error("BR handling failed")
	}
}

func TestReadSmiNbsp(t *testing.T) {
	b, err := ReadSmi(bytes.NewReader([]byte(t3.raw)))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b, t3.exp) {
		t.Log("got:", b)
		t.Log("exp:", t3.exp)
		t.Error("&nbsp; handling failed")
	}
}

func TestReadSmiSync(t *testing.T) {
	b, err := ReadSmi(bytes.NewReader([]byte(t4.raw)))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b, t4.exp) {
		t.Log("got:", b)
		t.Log("exp:", t4.exp)
		t.Error("Sync handling failed")
	}
}
