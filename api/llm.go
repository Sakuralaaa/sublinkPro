package api

import (
	"sublink/node/protocol"
	"sublink/services/llm"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// LLMOrganizeNodes 使用LLM整理节点
func LLMOrganizeNodes(c *gin.Context) {
	var req struct {
		Nodes []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Link     string `json:"link"`
			Country  string `json:"country"`
			Group    string `json:"group"`
		} `json:"nodes"`
		Instruction string `json:"instruction"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	if len(req.Nodes) == 0 {
		utils.FailWithMsg(c, "节点列表不能为空")
		return
	}

	// Convert to LLM node info (only send non-sensitive info)
	nodes := make([]llm.NodeInfo, 0, len(req.Nodes))
	for _, n := range req.Nodes {
		nodes = append(nodes, llm.NodeInfo{
			ID:       n.ID,
			Name:     n.Name,
			Protocol: protocol.GetProtocolFromLink(n.Link),
			Country:  n.Country,
			Group:    n.Group,
		})
	}

	result, err := llm.OrganizeNodes(nodes, req.Instruction)
	if err != nil {
		utils.FailWithMsg(c, "LLM整理失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "整理完成", gin.H{
		"result": result,
	})
}

// LLMGenerateRules 使用LLM生成订阅规则
func LLMGenerateRules(c *gin.Context) {
	var req struct {
		Nodes []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Link     string `json:"link"`
			Country  string `json:"country"`
			Group    string `json:"group"`
		} `json:"nodes"`
		ClientType  string `json:"clientType"`
		Instruction string `json:"instruction"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	if len(req.Nodes) == 0 {
		utils.FailWithMsg(c, "节点列表不能为空")
		return
	}

	if req.ClientType == "" {
		req.ClientType = "clash"
	}

	// Convert to LLM node info (only send non-sensitive info)
	nodes := make([]llm.NodeInfo, 0, len(req.Nodes))
	for _, n := range req.Nodes {
		nodes = append(nodes, llm.NodeInfo{
			ID:       n.ID,
			Name:     n.Name,
			Protocol: protocol.GetProtocolFromLink(n.Link),
			Country:  n.Country,
			Group:    n.Group,
		})
	}

	result, err := llm.GenerateRules(nodes, req.ClientType, req.Instruction)
	if err != nil {
		utils.FailWithMsg(c, "规则生成失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "生成完成", gin.H{
		"result": result,
	})
}

// LLMTestConnection 测试LLM API连接
func LLMTestConnection(c *gin.Context) {
	var req struct {
		APIUrl string `json:"apiUrl"`
		APIKey string `json:"apiKey"`
		Model  string `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	if req.APIUrl == "" || req.APIKey == "" {
		utils.FailWithMsg(c, "API URL 和 API Key 不能为空")
		return
	}

	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}

	config := &llm.LLMConfig{
		APIUrl: req.APIUrl,
		APIKey: req.APIKey,
		Model:  req.Model,
	}

	if err := llm.TestConnection(config); err != nil {
		utils.FailWithMsg(c, "连接测试失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "连接测试成功")
}
