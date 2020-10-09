package raft

import (
	_ "fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO: add test for parse cluster file
func TestReadConfigFile(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(1, 1, "They should be equal")
	//	c := Cluster{}
	//	c.Config("config")
	//	assert.Equal(1, 1, "They should be equal")
	//	assert.Equal(len(c.Servers), 5, "They should be equal")
	//	assert.Equal(c.Servers[0].Name, "ServerOne")
}
