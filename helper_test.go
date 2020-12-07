package gohealthchecker

import "testing"

func TestHelperGetHostIpAddress(t *testing.T) {
	info := &SystemInformation{}
	myIp := info.getIpAddress()

	if myIp.ipAddress == "" {
		t.Errorf("ip address should not be empty")
	}

	if myIp.ipAddress == "127.0.0.1" || myIp.ipAddress == "0.0.0.0" {
		t.Errorf("ip address should not be localhost")
	}
}

func TestHelperGetRuntimeVersion(t *testing.T) {
	info := &SystemInformation{}
	version := info.getRuntimeVersion()

	if version == "" {
		t.Errorf("version should not be empty")
	}
}

func TestGetStatusSystem(t *testing.T) {
	info := &SystemInformation{}
	if err := info.GetSystemInfo(); err != nil {
		t.Errorf(err.Error())
	}
}
