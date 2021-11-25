package main

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"testing"
)

func TestExecute(t *testing.T) {
	// A Controller represents the top-level control of a mock ecosystem.
	ctrl := gomock.NewController(t)
	// Create the mocks.
	mockIpGetter := NewMockIPGetter(ctrl)
	mockFileWriter := NewMockFileWriter(ctrl)

	expectedOutputFile := "output.txt"
	expectedIP := "184.162.7.66"

	// I expect GetPublicIp to return the IP above.
	mockIpGetter.EXPECT().
		GetPublicIP().
		Return(&expectedIP, nil).
		Times(1)

	// I expect this ip to be written in the file output.txt.
	mockFileWriter.EXPECT().
		Write(expectedOutputFile, []byte(expectedIP), fs.FileMode(0644)).
		Return(nil).
		Times(1)

	// Run the code.
	err := Execute(mockIpGetter, mockFileWriter, expectedOutputFile)

	// Ensure there are no errors for this scenario.
	assert.NoError(t, err, "no errors expected")

	//Ensure that all the EXPECTed mocks have been called.
	ctrl.Finish()
}
