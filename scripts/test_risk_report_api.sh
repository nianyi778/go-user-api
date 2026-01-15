#!/bin/bash

# 风险报告使用记录 API 测试脚本
# 使用方法: ./test_risk_report_api.sh

# 配置
API_BASE_URL="http://localhost:8080/api/v1"
API_KEY="dev-test-key-please-change-in-production"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}风险报告使用记录 API 测试${NC}"
echo -e "${YELLOW}========================================${NC}\n"

# 测试1: 创建单条使用记录
echo -e "${YELLOW}测试1: 创建单条使用记录${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
  "${API_BASE_URL}/risk-report/usage" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{
    "user_id": "123456789",
    "ticker": "AAPL",
    "request_time": "2026-01-15T02:30:45Z",
    "response_time": "2026-01-15T02:30:52Z",
    "prompt_tokens": 1872,
    "completion_tokens": 580,
    "total_tokens": 2452,
    "ai_response": "【当前阶段判断】★★★★★\n明确判断：高位回调",
    "stock_price": 259.83,
    "market_state": "PRE",
    "news_sentiment_score": 90,
    "news_sentiment_label": "偏多",
    "peak_signals_triggered": 1,
    "action_suggestion": "偏买入/试探",
    "rate_limit_remaining": 8,
    "response_duration_ms": 7311
  }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 201 ]; then
  echo -e "${GREEN}✓ 成功${NC}"
  echo "$body" | jq '.'
  record_id=$(echo "$body" | jq -r '.data.record_id')
else
  echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
  echo "$body"
fi
echo ""

# 测试2: 无效的 API Key
echo -e "${YELLOW}测试2: 测试无效的 API Key${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
  "${API_BASE_URL}/risk-report/usage" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: invalid-key" \
  -d '{
    "user_id": "123456789",
    "ticker": "AAPL",
    "request_time": "2026-01-15T02:30:45Z",
    "response_time": "2026-01-15T02:30:52Z",
    "prompt_tokens": 100,
    "completion_tokens": 50,
    "total_tokens": 150,
    "ai_response": "test"
  }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 401 ]; then
  echo -e "${GREEN}✓ 正确拦截${NC}"
  echo "$body" | jq '.'
else
  echo -e "${RED}✗ 应该返回 401${NC}"
  echo "$body"
fi
echo ""

# 测试3: 验证错误 - token 数量不匹配
echo -e "${YELLOW}测试3: 验证错误 - token 数量不匹配${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
  "${API_BASE_URL}/risk-report/usage" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{
    "user_id": "123456789",
    "ticker": "AAPL",
    "request_time": "2026-01-15T02:30:45Z",
    "response_time": "2026-01-15T02:30:52Z",
    "prompt_tokens": 100,
    "completion_tokens": 50,
    "total_tokens": 200,
    "ai_response": "test"
  }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 400 ]; then
  echo -e "${GREEN}✓ 正确验证${NC}"
  echo "$body" | jq '.'
else
  echo -e "${RED}✗ 应该返回 400${NC}"
  echo "$body"
fi
echo ""

# 测试4: 批量创建
echo -e "${YELLOW}测试4: 批量创建使用记录${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
  "${API_BASE_URL}/risk-report/usage/batch" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{
    "records": [
      {
        "user_id": "111111111",
        "ticker": "TSLA",
        "request_time": "2026-01-15T03:00:00Z",
        "response_time": "2026-01-15T03:00:05Z",
        "prompt_tokens": 1000,
        "completion_tokens": 500,
        "total_tokens": 1500,
        "ai_response": "特斯拉分析..."
      },
      {
        "user_id": "222222222",
        "ticker": "NVDA",
        "request_time": "2026-01-15T04:00:00Z",
        "response_time": "2026-01-15T04:00:06Z",
        "prompt_tokens": 1200,
        "completion_tokens": 600,
        "total_tokens": 1800,
        "ai_response": "英伟达分析..."
      }
    ]
  }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 200 ]; then
  echo -e "${GREEN}✓ 成功${NC}"
  echo "$body" | jq '.'
else
  echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
  echo "$body"
fi
echo ""

# 测试5: 查询记录列表
if [ -n "$record_id" ]; then
  echo -e "${YELLOW}测试5: 查询记录详情${NC}"
  response=$(curl -s -w "\n%{http_code}" -X GET \
    "${API_BASE_URL}/risk-report/usage/${record_id}" \
    -H "X-API-Key: ${API_KEY}")

  http_code=$(echo "$response" | tail -n1)
  body=$(echo "$response" | sed '$d')

  if [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✓ 成功${NC}"
    echo "$body" | jq '.data | {id, user_id, ticker, total_tokens}'
  else
    echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
    echo "$body"
  fi
  echo ""
fi

# 测试6: 查询用户统计
echo -e "${YELLOW}测试6: 查询用户统计信息${NC}"
response=$(curl -s -w "\n%{http_code}" -X GET \
  "${API_BASE_URL}/risk-report/usage/stats/123456789" \
  -H "X-API-Key: ${API_KEY}")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 200 ]; then
  echo -e "${GREEN}✓ 成功${NC}"
  echo "$body" | jq '.'
else
  echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
  echo "$body"
fi
echo ""

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}测试完成${NC}"
echo -e "${YELLOW}========================================${NC}"
