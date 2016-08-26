package cloudfoundry

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/nu7hatch/gouuid"
)

type TaskGuid struct {
	AppGuid *uuid.UUID
	TaskId  []byte
}

func NewTaskGuid(taskGuid string) (TaskGuid, error) {
	if len(taskGuid) < 36 {
		return TaskGuid{}, errors.New("invalid task guid")
	}

	appGuid, err := uuid.ParseHex(taskGuid[:36])
	if err != nil {
		return TaskGuid{}, err
	}

	taskId, err := hex.DecodeString(taskGuid[37:])
	if err != nil {
		return TaskGuid{}, err
	}

	return TaskGuid{
		AppGuid: appGuid,
		TaskId:  taskId,
	}, nil
}

func (pg TaskGuid) ShortenedGuid() string {
	shortAppGuid := trimPadding(base32.StdEncoding.EncodeToString(pg.AppGuid[:]))
	shortTaskId := trimPadding(base32.StdEncoding.EncodeToString(pg.TaskId))

	return strings.ToLower(shortAppGuid + "-" + shortTaskId)
}

func (pg TaskGuid) String() string {
	return pg.AppGuid.String() + "-" + hex.EncodeToString(pg.TaskId)
}

func DecodeTaskGuid(shortenedGuid string) (TaskGuid, error) {
	splited := strings.Split(strings.ToUpper(shortenedGuid), "-")
	if len(splited) != 2 {
		return TaskGuid{}, errors.New("invalid-shortened-guid")
	}
	// add padding
	appGuid := addPadding(splited[0])
	taskId := addPadding(splited[1])

	// decode it
	longAppGuid, err := base32.StdEncoding.DecodeString(appGuid[:])
	if err != nil {
		return TaskGuid{}, err
	}
	appGuidUUID, err := uuid.Parse(longAppGuid)
	if err != nil {
		return TaskGuid{}, err
	}

	// decode it
	longTaskId, err := base32.StdEncoding.DecodeString(taskId[:])
	if err != nil {
		return TaskGuid{}, err
	}

	return TaskGuid{
		AppGuid: appGuidUUID,
		TaskId:  longTaskId,
	}, nil
}
