package common

import (
	"errors"
	"fmt"
	"strconv"
)

type DifficultyTarget uint32

func (t DifficultyTarget) UInt32() uint32 {
	return uint32(t)
}

func (t *DifficultyTarget) Set(s string) error {
	target64, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return errors.New("must fit into uint32")
	}
	*t = DifficultyTarget(uint32(target64))
	return nil
}

func (t DifficultyTarget) String() string {
	return fmt.Sprintf("%d", t)
}
