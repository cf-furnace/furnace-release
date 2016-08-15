package cloudfoundry

import (
	"encoding/base32"
	"errors"
	"strings"

	"github.com/nu7hatch/gouuid"
)

type ProcessGuid struct {
	AppGuid    *uuid.UUID
	AppVersion *uuid.UUID
}

func NewProcessGuid(processGuid string) (ProcessGuid, error) {
	if len(processGuid) < 36 {
		return ProcessGuid{}, errors.New("invalid process guid")
	}

	appGuid, err := uuid.ParseHex(processGuid[:36])
	if err != nil {
		return ProcessGuid{}, err
	}

	appVersion, err := uuid.ParseHex(processGuid[37:])
	if err != nil {
		return ProcessGuid{}, err
	}

	return ProcessGuid{
		AppGuid:    appGuid,
		AppVersion: appVersion,
	}, nil
}

func (pg ProcessGuid) ShortenedGuid() string {
	shortAppGuid := trimPadding(base32.StdEncoding.EncodeToString(pg.AppGuid[:]))
	shortAppVersion := trimPadding(base32.StdEncoding.EncodeToString(pg.AppVersion[:]))

	return strings.ToLower(shortAppGuid + "-" + shortAppVersion)
}

func (pg ProcessGuid) String() string {
	return pg.AppGuid.String() + "-" + pg.AppVersion.String()
}

func DecodeProcessGuid(shortenedGuid string) (ProcessGuid, error) {
	splited := strings.Split(strings.ToUpper(shortenedGuid), "-")
	if len(splited) != 2 {
		return ProcessGuid{}, errors.New("invalid-shortened-guid")
	}

	// add padding
	appGuid := addPadding(splited[0])
	appVersion := addPadding(splited[1])

	// decode it
	longAppGuid, err := base32.StdEncoding.DecodeString(appGuid[:])
	if err != nil {
		return ProcessGuid{}, err
	}
	appGuidUUID, err := uuid.Parse(longAppGuid)
	if err != nil {
		return ProcessGuid{}, err
	}

	// decode it
	longAppVersion, err := base32.StdEncoding.DecodeString(appVersion[:])
	if err != nil {
		return ProcessGuid{}, err
	}
	appVersionUUID, err := uuid.Parse(longAppVersion)
	if err != nil {
		return ProcessGuid{}, err
	}

	return ProcessGuid{
		AppGuid:    appGuidUUID,
		AppVersion: appVersionUUID,
	}, nil
}

func addPadding(s string) string {
	return s + strings.Repeat("=", 8-len(s)%8)
}

func trimPadding(s string) string {
	return strings.TrimRight(s, "=")
}
