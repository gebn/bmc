package ipmi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestFullSensorRecordDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *FullSensorRecord
	}{
		{
			[]byte{
				// key
				0x20, // owned by the BMC
				0x00, // channel 0, system software owns sensor
				0x01, // sensor number 1

				// body
				0x03, // processor entity ID
				0x01, // treat as physical entity, instance number 1, system-relative
				0x7f, // not settable, scanning, event generation, thresholds, hysteresis and sensor event type/reading type code initialised, event generation and scanning enabled on power-up
				0x68, // don't ignore sensor if Entity is not present or disabled, sensor automatically rearms itself when the event clears, hysteresis is readable and settable, thresholds are readable and settable per Reading Mask and Settable Threshold Mask respectively, per threshold/discrete-state event enable/disable control (implies that entire sensor and global disable are also supported). Basically, it can do everything.
				0x01, // sensor type 1 (Temperature)
				0x01, // Event / Reading Type Code 1
				0x00, 0x72,
				0x00, 0x72,
				0x3f, 0x3f,
				0x80,                                           // units 1: 2’s complement (signed), no rate unit, no modifier unit, not a percentage
				0x01,                                           // units 2: base unit is degrees C
				0x00,                                           // units 3: modifier unit is unused
				0x00,                                           // linearisation: linear
				0x01,                                           // LS 8 bits of M, which is a signed 10-bit 2's complement value
				0x00,                                           // MS 2 bits of M (can now see M = 1), tolerance: 0
				0x00,                                           // LS 8 bits of B, which is a signed 10-bit 2's complement value
				0x00,                                           // MS 2 bits of B (can now see B = 0), LS 6 bits of accuracy
				0x00,                                           // MS 4 bits of accuracy (can now see accuracy = 0), accuracy exp: 0, sensor direction N/A
				0x00,                                           // R exp: 0, B exp: 0
				0x07,                                           // analogue characteristic flags: normal min and max specified, nominal reading specified
				0x28,                                           // nominal reading *raw* value -> (1*0x28 + (0*10^0))*10^0 = 40 celsius, // N.B. 2's complement due to units 1
				0x59,                                           // nominal max *raw* value -> (1*0x59 + (0*10^0))*10^0 = 89 celsius, // N.B. 2's complement due to units 1
				0xfc,                                           // nominal min *raw* value -> (1*0xfc + (0*10^0))*10^0 = -4 celsius, // N.B. 2's complement due to units 1
				0x7f,                                           // sensor max reading *raw* value -> (1*0x7f + (0*10^0))*10^0 = 127 celsius, // N.B. 2's complement due to units 1
				0x80,                                           // sensor min reading *raw* value -> (1*0x80 + (0*10^0))*10^0 = -128 celsius, // N.B. 2's complement due to units 1
				0x64, 0x64, 0x5f, 0x00, 0x00, 0x00, 0x02, 0x02, // thresholds
				0x00, 0x00, 0x00, // reserved
				0xc8, // 8-bit ASCII + Latin 1, followed by 8 chars (takes to end of packet)
				0x43, // C
				0x50, // P
				0x55, // U
				0x20, // <space>
				0x54, // T
				0x65, // e
				0x6d, // m
				0x70, // p
			},
			&FullSensorRecord{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x20, 0x00, 0x01, 0x03, 0x01, 0x7f, 0x68, 0x01, 0x01,
						0x00, 0x72, 0x00, 0x72, 0x3f, 0x3f, 0x80, 0x01, 0x00,
						0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x28,
						0x59, 0xfc, 0x7f, 0x80, 0x64, 0x64, 0x5f, 0x00, 0x00,
						0x00, 0x02, 0x02, 0x00, 0x00, 0x00, 0xc8, 0x43, 0x50,
						0x55, 0x20, 0x54, 0x65, 0x6d, 0x70,
					},
					Payload: []byte{},
				},
				SensorRecordKey: SensorRecordKey{
					OwnerAddress: SlaveAddressBMC.Address(),
					Channel:      ChannelPrimaryIPMB,
					OwnerLUN:     LUNBMC,
					Number:       1,
				},
				ConversionFactors: ConversionFactors{
					M:    1,
					B:    0,
					BExp: 0,
					RExp: 0,
				},
				Entity:                  EntityIDProcessor,
				IsContainerEntity:       false,
				Instance:                1,
				Ignore:                  false,
				SensorType:              SensorTypeTemperature,
				OutputType:              OutputTypeThreshold,
				AnalogDataFormat:        AnalogDataFormatTwosComplement,
				RateUnit:                RateUnitNone,
				IsPercentage:            false,
				BaseUnit:                SensorUnitCelsius,
				ModifierUnit:            0,
				Linearisation:           LinearisationLinear,
				Tolerance:               0,
				Accuracy:                0,
				AccuracyExp:             0,
				Direction:               SensorDirectionUnspecified,
				NominalReadingSpecified: true,
				NormalMinSpecified:      true,
				NormalMaxSpecified:      true,
				NominalReading:          0x28,
				NormalMin:               0xfc,
				NormalMax:               0x59,
				SensorMin:               0x80,
				SensorMax:               0x7f,
				Identity:                "CPU Temp",
			},
		},
		{
			[]byte{
				// key
				0x30, // owner
				0x5e, // channel 5, LUN 2
				0x16, // sensor number 22

				// body
				0x0a, // PSU entity ID
				0xe0, // treat as logical entity, instance number 96, device-relative
				0x7f, // not settable, scanning, event generation, thresholds, hysteresis and sensor event type/reading type code initialised, event generation and scanning enabled on power-up
				0xe8, // ignore sensor if Entity is not present or disabled, sensor automatically rearms itself when the event clears, hysteresis is readable and settable, thresholds are readable and settable per Reading Mask and Settable Threshold Mask respectively, per threshold/discrete-state event enable/disable control (implies that entire sensor and global disable are also supported). Basically, it can do everything.
				0x03, // sensor type 3 (Current)
				0x01, // Event / Reading Type Code 1
				0x00, 0x72,
				0x00, 0x72,
				0x3f, 0x3f,
				0b00101101,                                     // units 1: unsigned, per hour, multiplication modifier, is a percentage
				0x05,                                           // units 2: base unit is amps
				0x0e,                                           // units 3: modifier unit is kilopascals
				0x05,                                           // linearisation: exp10
				0xff,                                           // LS 8 bits of M, which is a signed 10-bit 2's complement value
				0b10110101,                                     // MS 2 bits of M (can now see M = -257), tolerance: 53
				0xf0,                                           // LS 8 bits of B, which is a signed 10-bit 2's complement value
				0x6a,                                           // MS 2 bits of B (can now see B = 496), LS 6 bits of accuracy
				0xad,                                           // MS 4 bits of accuracy (can now see accuracy = -342), accuracy exp: 3, sensor direction input
				0b10100101,                                     // R exp: -6, B exp: 5
				0xaa,                                           // analogue characteristic flags: normal min unspecified, max specified, nominal reading unspecified
				0x08,                                           // nominal reading raw, unsigned
				0x11,                                           // nominal max raw, unsigned
				0x3a,                                           // nominal min raw, unsigned
				0x7b,                                           // sensor max reading raw, unsigned
				0x80,                                           // sensor min reading raw, unsigned
				0x64, 0x64, 0x5f, 0x00, 0x00, 0x00, 0x02, 0x02, // thresholds
				0xff, 0xff, 0xff, // reserved
				0x89, // 6-bit ASCII, followed by 9 chars
				// 8	0x18	0b011000
				// $	0x04	0b000100
				// 		0x00	0b000000
				// =	0x1d	0b011101
				0b00011000,
				0b00000001,
				0b01110100,

				// '	0x07	0b000111
				// [	0x3b	0b111011
				// \	0x3c	0b111100
				// V	0x36	0b110110
				0b11000111,
				0b11001110,
				0b11011011,

				// _	0x3f	0b111111
				0b00111111,
				0x9a, // 3 bytes of trailing data
				0x00,
				0x00,
			},
			&FullSensorRecord{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x30, 0x5e, 0x16, 0x0a, 0xe0, 0x7f, 0xe8, 0x03, 0x01,
						0x00, 0x72, 0x00, 0x72, 0x3f, 0x3f, 0x2d, 0x05, 0x0e,
						0x05, 0xff, 0xb5, 0xf0, 0x6a, 0xad, 0xa5, 0xaa, 0x08,
						0x11, 0x3a, 0x7b, 0x80, 0x64, 0x64, 0x5f, 0x00, 0x00,
						0x00, 0x02, 0x02, 0xff, 0xff, 0xff, 0x89, 0x18, 0x01,
						0x74, 0xc7, 0xce, 0xdb, 0x3f,
					},
					Payload: []byte{0x9a, 0x00, 0x00},
				},
				SensorRecordKey: SensorRecordKey{
					OwnerAddress: SlaveAddress(24).Address(),
					Channel:      Channel(5),
					OwnerLUN:     LUNSMS,
					Number:       22,
				},
				ConversionFactors: ConversionFactors{
					M:    -257,
					B:    496,
					BExp: 5,
					RExp: -6,
				},
				Entity:                  EntityIDPowerSupply,
				IsContainerEntity:       true,
				Instance:                96,
				Ignore:                  true,
				SensorType:              SensorTypeCurrent,
				OutputType:              OutputTypeThreshold,
				AnalogDataFormat:        AnalogDataFormatUnsigned,
				RateUnit:                RateUnitPerHour,
				IsPercentage:            true,
				BaseUnit:                SensorUnitAmps,
				ModifierUnit:            SensorUnitKilopascals,
				Linearisation:           LinearisationExp10,
				Tolerance:               53,
				Accuracy:                -342,
				AccuracyExp:             3,
				Direction:               SensorDirectionInput,
				NominalReadingSpecified: false,
				NormalMinSpecified:      false,
				NormalMaxSpecified:      true,
				NominalReading:          0x08,
				NormalMin:               0x3a,
				NormalMax:               0x11,
				SensorMin:               0x80,
				SensorMax:               0x7b,
				Identity:                `8$ ='[\V_`,
			},
		},
	}
	for _, test := range tests {
		fsr := &FullSensorRecord{}
		err := fsr.DecodeFromBytes(test.in, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("expected error decoding %v, got none", test.in)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, fsr); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, fsr, test.want, diff)
			}
		case err != nil && test.want != nil:
			t.Errorf("unexpected error: %v", err)
		}
	}
}
