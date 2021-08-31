package common

import "testing"

func TestFindMountPoint(t *testing.T) {
	res, err := FindMountPoint()
	Must(err)
	t.Logf("Mount info :%v", res)
}
