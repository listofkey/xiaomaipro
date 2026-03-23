package common

import (
	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	snowflakeNode *snowflake.Node
)

func init() {
	node, err := snowflake.NewNode(2)
	if err != nil {
		logx.Errorf("初始化雪花算法节点失败: %v", err)
		panic(err)
	}
	snowflakeNode = node
}
func GenerateId() int64 {
	return snowflakeNode.Generate().Int64()
}
