/*
Copyright (c) Facebook, Inc. and its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package chrony

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

/*
The unittests here contain packets in binary form.
The easiest way to obtain those is using tcpdump/tshark.

For example, running `tshark -i any -T fields -e data.data udp and port 323` in one shell and
using `chronyc` cli in another allows to capture all the bytes sent and received.

Alternatively, strace can be used:
`strace -xx -e sendto,recvfrom -v -s 10000 chronyc sources` will print sent and received bytes.
Using strace is the only option when working with private parts of the
chrony protocol (commands that only work over the unix socket), like `chronyc ntpdata`.
*/

func TestDecodeUnauthorized(t *testing.T) {
	raw := []uint8{
		0x06, 0x02, 0x00, 0x00, 0x00, 0x39, 0x00, 0x01, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa3, 0xa8, 0xc8, 0x40,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	_, err := decodePacket(raw)
	require.Error(t, err)
}

func TestDecodeSources(t *testing.T) {
	raw := []uint8{
		0x06, 0x02, 0x00, 0x00, 0x00, 0x0e, 0x00, 0x02, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x39, 0x3a, 0xb1, 0x23,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x12,
	}
	packet, err := decodePacket(raw)
	require.Nil(t, err)
	want := &ReplySources{
		ReplyHead: ReplyHead{
			Version:  protoVersionNumber,
			PKTType:  pktTypeCmdReply,
			Command:  reqNSources,
			Reply:    rpyNSources,
			Status:   sttSuccess,
			Sequence: 960147747,
		},
		NSources: 18,
	}
	require.Equal(t, want, packet)
}

func TestDecodeSourceData(t *testing.T) {
	raw := []uint8{
		0x06, 0x02, 0x00, 0x00, 0x00, 0x0f, 0x00, 0x03, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x83, 0xbf, 0x73,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x24, 0x01,
		0xdb, 0x00, 0x31, 0x10, 0x20, 0xc0, 0xfa, 0xce, 0x00, 0x00,
		0x00, 0x48, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x0a,
		0x00, 0x02, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff,
		0x00, 0x00, 0x06, 0xa9, 0xe6, 0xc5, 0xee, 0xf3, 0xe6, 0xd1,
		0x4f, 0xbe, 0xea, 0xbb, 0x92, 0x3b,
	}
	packet, err := decodePacket(raw)
	require.Nil(t, err)
	want := &ReplySourceData{
		ReplyHead: ReplyHead{
			Version:  protoVersionNumber,
			PKTType:  pktTypeCmdReply,
			Command:  reqSourceData,
			Reply:    rpySourceData,
			Status:   sttSuccess,
			Sequence: 209960819,
		},
		SourceData: SourceData{
			IPAddr:         net.IP{0x24, 0x01, 0xdb, 0x00, 0x31, 0x10, 0x20, 0xc0, 0xfa, 0xce, 0x00, 0x00, 0x00, 0x48, 0x00, 0x00},
			Poll:           10,
			Stratum:        2,
			State:          4,
			Mode:           0,
			Flags:          0,
			Reachability:   255,
			SinceSample:    1705,
			OrigLatestMeas: 4.719099888461642e-05,
			LatestMeas:     4.990374873159453e-05,
			LatestMeasErr:  0.00017888184811454266,
		},
	}
	require.Equal(t, want, packet)
}

func TestDecodeSourceStats(t *testing.T) {
	raw := []uint8{
		0x06, 0x02, 0x00, 0x00, 0x00, 0x22, 0x00, 0x06, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x59, 0x95, 0xd8, 0xfa,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xbf, 0x8b,
		0xe5, 0xe9, 0x24, 0x01, 0xdb, 0x00, 0x31, 0x10, 0x20, 0xc0,
		0xfa, 0xce, 0x00, 0x00, 0x00, 0x48, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x05,
		0x00, 0x00, 0x1a, 0x27, 0xe4, 0x94, 0x84, 0x99, 0xed, 0x34,
		0xe0, 0x09, 0xf6, 0xc0, 0x64, 0x94, 0xdf, 0x18, 0xb4, 0x76,
		0xea, 0xb9, 0xc0, 0xa1,
	}
	packet, err := decodePacket(raw)
	require.Nil(t, err)
	want := &ReplySourceStats{
		ReplyHead: ReplyHead{
			Version:  protoVersionNumber,
			PKTType:  pktTypeCmdReply,
			Command:  reqSourceStats,
			Reply:    rpySourceStats,
			Status:   sttSuccess,
			Sequence: 1502992634,
		},
		SourceStats: SourceStats{
			RefID:              3213616617,
			IPAddr:             net.IP{0x24, 0x01, 0xdb, 0x00, 0x31, 0x10, 0x20, 0xc0, 0xfa, 0xce, 0x00, 0x00, 0x00, 0x48, 0x00, 0x00},
			NSamples:           12,
			NRuns:              5,
			SpanSeconds:        6695,
			StandardDeviation:  1.770472044881899e-05,
			ResidFreqPPM:       -0.00038742992910556495,
			SkewPPM:            0.0117427296936512,
			EstimatedOffset:    -3.44656518791453e-06,
			EstimatedOffsetErr: 0.0001771473471308127,
		},
	}
	require.Equal(t, want, packet)
}

func TestDecodeTracking(t *testing.T) {
	raw := []uint8{
		0x06, 0x02, 0x00, 0x00, 0x00, 0x21, 0x00, 0x05, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xe6, 0x25,
		0xc6, 0x6e, 0x24, 0x01, 0xdb, 0x00, 0x31, 0x10, 0x21, 0x32,
		0xfa, 0xce, 0x00, 0x00, 0x00, 0x8e, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x61, 0x38, 0xe1, 0x81, 0x36, 0x94, 0x8d, 0xd5, 0xdf, 0x19,
		0x2d, 0xb7, 0xdf, 0x42, 0x83, 0xf5, 0xe2, 0xeb, 0xca, 0x12,
		0x05, 0x39, 0xe1, 0x11, 0xeb, 0x7b, 0x3e, 0x5d, 0xf4, 0xb0,
		0x75, 0x12, 0xea, 0xe7, 0x5b, 0x0c, 0xf0, 0x88, 0x1d, 0x4e,
		0x16, 0x82, 0x1f, 0x69,
	}
	packet, err := decodePacket(raw)
	require.Nil(t, err)
	want := &ReplyTracking{
		ReplyHead: ReplyHead{
			Version:  protoVersionNumber,
			PKTType:  pktTypeCmdReply,
			Res1:     0,
			Res2:     0,
			Command:  reqTracking,
			Reply:    rpyTracking,
			Status:   sttSuccess,
			Sequence: 2,
		},
		Tracking: Tracking{
			RefID:              3861235310,
			IPAddr:             net.IP{36, 1, 219, 0, 49, 16, 33, 50, 250, 206, 0, 0, 0, 142, 0, 0},
			Stratum:            3,
			LeapStatus:         0,
			RefTime:            time.Unix(0, 1631117697915705301),
			CurrentCorrection:  -3.4395072816550964e-06,
			LastOffset:         -2.823539716700907e-06,
			RMSOffset:          1.405413968313951e-05,
			FreqPPM:            -1.5478190183639526,
			ResidFreqPPM:       -0.00012660636275541037,
			SkewPPM:            0.005385049618780613,
			RootDelay:          0.00022063794312998652,
			RootDispersion:     0.0010384710039943457,
			LastUpdateInterval: 520.4907836914062,
		},
	}
	require.Equal(t, want, packet)
}

/* private part of the protocol */

func TestDecodeServerStats(t *testing.T) {
	raw := []uint8{
		0x06, 0x02, 0x00, 0x00, 0x00, 0x36, 0x00, 0x0e, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x07, 0x16, 0xff,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x10, 0x03, 0xcd, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	packet, err := decodePacket(raw)
	require.Nil(t, err)
	want := &ReplyServerStats{
		ReplyHead: ReplyHead{
			Version:  protoVersionNumber,
			PKTType:  pktTypeCmdReply,
			Res1:     0,
			Res2:     0,
			Command:  reqServerStats,
			Reply:    rpyServerStats,
			Status:   sttSuccess,
			Sequence: 50796287,
		},
		ServerStats: ServerStats{
			NTPHits:  0,
			CMDHits:  1049549,
			NTPDrops: 0,
			CMDDrops: 0,
			LogDrops: 0,
		},
	}
	require.Equal(t, want, packet)
}

func TestDecodeServerStats2(t *testing.T) {
	raw := []uint8{
		0x06, 0x02, 0x00, 0x00, 0x00, 0x36, 0x00, 0x16, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x07, 0x16, 0xff,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x10, 0x03, 0xcd, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x21, 0x00,
		0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	packet, err := decodePacket(raw)
	require.Nil(t, err)
	want := &ReplyServerStats2{
		ReplyHead: ReplyHead{
			Version:  protoVersionNumber,
			PKTType:  pktTypeCmdReply,
			Res1:     0,
			Res2:     0,
			Command:  reqServerStats,
			Reply:    rpyServerStats2,
			Status:   sttSuccess,
			Sequence: 50796287,
		},
		ServerStats2: ServerStats2{
			NTPHits:     0,
			NKEHits:     1049549,
			CMDHits:     0,
			NTPDrops:    0,
			NKEDrops:    0,
			CMDDrops:    553648383,
			LogDrops:    0,
			NTPAuthHits: 0,
		},
	}
	require.Equal(t, want, packet)
}

func TestDecodeNTPData(t *testing.T) {
	raw := []uint8{
		0x06, 0x02, 0x00, 0x00, 0x00, 0x39, 0x00, 0x10, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xe9, 0xb2, 0x80, 0xdb,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x24, 0x01,
		0xdb, 0x00, 0x23, 0x1c, 0x28, 0x12, 0xfa, 0xce, 0x00, 0x00,
		0x01, 0x7b, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x24, 0x01,
		0xdb, 0x00, 0xee, 0xf0, 0x11, 0x20, 0x35, 0x20, 0x00, 0x00,
		0x20, 0x08, 0x0f, 0x06, 0x00, 0x02, 0x00, 0x00, 0x00, 0x7b,
		0x00, 0x04, 0x04, 0x02, 0x0a, 0xe8, 0xe4, 0x80, 0x00, 0x00,
		0xe4, 0x80, 0x00, 0x00, 0x23, 0xe1, 0x0b, 0x36, 0x00, 0x00,
		0x00, 0x00, 0x61, 0x3a, 0x39, 0xf0, 0x06, 0x6a, 0xe1, 0xf8,
		0xf3, 0x50, 0x79, 0x73, 0xfc, 0xa1, 0x7d, 0x6e, 0xd4, 0xb6,
		0x81, 0xb7, 0xe6, 0xd1, 0xb9, 0x3d, 0x01, 0x04, 0xb6, 0xad,
		0x43, 0xfd, 0x4b, 0x4b, 0x00, 0x00, 0x11, 0x2f, 0x00, 0x00,
		0x11, 0x2c, 0x00, 0x00, 0x11, 0x2c, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff,
	}
	packet, err := decodePacket(raw)
	require.Nil(t, err)
	want := &ReplyNTPData{
		ReplyHead: ReplyHead{
			Version:  protoVersionNumber,
			PKTType:  pktTypeCmdReply,
			Res1:     0,
			Res2:     0,
			Command:  reqNTPData,
			Reply:    rpyNTPData,
			Status:   sttSuccess,
			Sequence: 3920789723,
		},
		NTPData: NTPData{
			RemoteAddr:      net.IP{0x24, 0x01, 0xdb, 0x00, 0x23, 0x1c, 0x28, 0x12, 0xfa, 0xce, 0x00, 0x00, 0x01, 0x7b, 0x00, 0x00},
			RemotePort:      123,
			LocalAddr:       net.IP{0x24, 0x01, 0xdb, 0x00, 0xee, 0xf0, 0x11, 0x20, 0x35, 0x20, 0x00, 0x00, 0x20, 0x08, 0x0f, 0x06},
			Leap:            0,
			Version:         4,
			Mode:            4,
			Stratum:         2,
			Poll:            10,
			Precision:       -24,
			RootDelay:       1.52587890625e-05,
			RootDispersion:  1.52587890625e-05,
			RefID:           601951030,
			RefTime:         time.Unix(0, 1631205872107667960),
			Offset:          -0.0026783079374581575,
			PeerDelay:       0.07885251939296722,
			PeerDispersion:  8.49863042162724e-08,
			ResponseTime:    5.000199962523766e-05,
			JitterAsymmetry: -0.49079379439353943,
			Flags:           17405,
			TXTssChar:       75,
			RXTssChar:       75,
			TotalTXCount:    4399,
			TotalRXCount:    4396,
			TotalValidCount: 4396,
		},
	}
	require.Equal(t, want, packet)
}