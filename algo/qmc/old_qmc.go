package qmc

import (
	"encoding/base64"
	"encoding/binary"
	"errors"

	"github.com/unlock-music/cli/algo/common"
)

var (
	ErrQmcFileLength      = errors.New("invalid qmc file length")
	ErrQmcKeyDecodeFailed = errors.New("base64 decode qmc key failed")
	ErrQmcKeyLength       = errors.New("unexpected decoded qmc key length")
)

type OldDecoder struct {
	file         []byte
	maskDetector func(encodedData []byte) (*Key256Mask, error)
	mask         *Key256Mask
	audioExt     string
	key          []byte
	audio        []byte
}

func NewMflac256Decoder(data []byte) common.Decoder {
	return &OldDecoder{file: data, maskDetector: detectMflac256Mask, audioExt: "flac"}
}

func NewMgg256Decoder(data []byte) common.Decoder {
	return &OldDecoder{file: data, maskDetector: detectMgg256Mask, audioExt: "ogg"}
}

func (d *OldDecoder) Validate() error {
	if nil != d.mask {
		return nil
	}
	if nil != d.maskDetector {
		if err := d.validateKey(); err != nil {
			return err
		}
		var err error
		d.mask, err = d.maskDetector(d.file)
		return err
	}
	return errors.New("no mask or mask detector found")
}

func (d *OldDecoder) validateKey() error {
	lenData := len(d.file)
	if lenData < 4 {
		return ErrQmcFileLength
	}

	keyLen := binary.LittleEndian.Uint32(d.file[lenData-4:])
	if lenData < int(keyLen+4) {
		return ErrQmcFileLength
	}
	var err error
	d.key, err = base64.StdEncoding.DecodeString(
		string(d.file[lenData-4-int(keyLen) : lenData-4]))
	if err != nil {
		return ErrQmcKeyDecodeFailed
	}

	if len(d.key) != 272 {
		return ErrQmcKeyLength
	}
	d.file = d.file[:lenData-4-int(keyLen)]
	return nil

}

func (d *OldDecoder) Decode() error {
	d.audio = d.mask.Decrypt(d.file)
	return nil
}

func (d OldDecoder) GetCoverImage() []byte {
	return nil
}

func (d OldDecoder) GetAudioData() []byte {
	return d.audio
}

func (d OldDecoder) GetAudioExt() string {
	if d.audioExt != "" {
		return "." + d.audioExt
	}
	return ""
}

func (d OldDecoder) GetMeta() common.Meta {
	return nil
}

func DecoderFuncWithExt(ext string) common.NewDecoderFunc {
	return func(file []byte) common.Decoder {
		return &OldDecoder{file: file, audioExt: ext, mask: getDefaultMask()}
	}
}

//goland:noinspection SpellCheckingInspection
func init() {
	common.RegisterDecoder("qmc0", false, DecoderFuncWithExt("mp3")) //QQ Music Mp3
	common.RegisterDecoder("qmc3", false, DecoderFuncWithExt("mp3")) //QQ Music Mp3

	common.RegisterDecoder("qmc2", false, DecoderFuncWithExt("m4a")) //QQ Music M4A
	common.RegisterDecoder("qmc4", false, DecoderFuncWithExt("m4a")) //QQ Music M4A
	common.RegisterDecoder("qmc6", false, DecoderFuncWithExt("m4a")) //QQ Music M4A
	common.RegisterDecoder("qmc8", false, DecoderFuncWithExt("m4a")) //QQ Music M4A

	common.RegisterDecoder("qmcflac", false, DecoderFuncWithExt("flac")) //QQ Music Flac
	common.RegisterDecoder("qmcogg", false, DecoderFuncWithExt("ogg"))   //QQ Music Ogg
	common.RegisterDecoder("tkm", false, DecoderFuncWithExt("m4a"))      //QQ Music Accompaniment M4a

	common.RegisterDecoder("bkcmp3", false, DecoderFuncWithExt("mp3"))   //Moo Music Mp3
	common.RegisterDecoder("bkcflac", false, DecoderFuncWithExt("flac")) //Moo Music Flac

	common.RegisterDecoder("666c6163", false, DecoderFuncWithExt("flac")) //QQ Music Weiyun Flac
	common.RegisterDecoder("6d7033", false, DecoderFuncWithExt("mp3"))    //QQ Music Weiyun Mp3
	common.RegisterDecoder("6f6767", false, DecoderFuncWithExt("ogg"))    //QQ Music Weiyun Ogg
	common.RegisterDecoder("6d3461", false, DecoderFuncWithExt("m4a"))    //QQ Music Weiyun M4a
	common.RegisterDecoder("776176", false, DecoderFuncWithExt("wav"))    //QQ Music Weiyun Wav

	common.RegisterDecoder("mgg", false, NewMgg256Decoder)     //QQ Music New Ogg
	common.RegisterDecoder("mflac", false, NewMflac256Decoder) //QQ Music New Flac
}