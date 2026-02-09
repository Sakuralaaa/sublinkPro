import request from './request';

// LLM 整理节点
export function llmOrganizeNodes(data) {
  return request({
    url: '/v1/llm/organize-nodes',
    method: 'post',
    data,
    timeout: 120000
  });
}

// LLM 生成订阅规则
export function llmGenerateRules(data) {
  return request({
    url: '/v1/llm/generate-rules',
    method: 'post',
    data,
    timeout: 120000
  });
}
