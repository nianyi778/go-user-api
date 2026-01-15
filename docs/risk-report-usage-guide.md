# Risk Report 使用记录功能使用指南

## 功能概述

本功能为 risk-report 项目提供使用记录上报接口，记录每次用户查询的详细信息，包括：
- 查询参数（用户ID、股票代码、时间）
- Token 消耗（prompt、completion、total）
- AI 响应内容
- 扩展信息（股价、市场状态、情绪分析等）

## API 接口文档

### 认证方式

所有 API 接口都需要在请求头中携带 API Key：

```bash
X-API-Key: your-api-key-here
```

### 接口列表

#### 1. 创建单条使用记录

**POST** `/api/v1/risk-report/usage`

请求示例：

```bash
curl -X POST http://localhost:8080/api/v1/risk-report/usage \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "user_id": "123456789",
    "ticker": "AAPL",
    "request_time": "2026-01-15T02:30:45Z",
    "response_time": "2026-01-15T02:30:52Z",
    "prompt_tokens": 1872,
    "completion_tokens": 580,
    "total_tokens": 2452,
    "ai_response": "【当前阶段判断】...",
    "stock_price": 259.83,
    "market_state": "PRE",
    "news_sentiment_score": 90,
    "news_sentiment_label": "偏多",
    "peak_signals_triggered": 1,
    "action_suggestion": "偏买入/试探",
    "rate_limit_remaining": 8,
    "response_duration_ms": 7311
  }'
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "message": "记录已保存",
    "record_id": "uuid-here"
  }
}
```

#### 2. 批量创建使用记录

**POST** `/api/v1/risk-report/usage/batch`

请求示例：

```bash
curl -X POST http://localhost:8080/api/v1/risk-report/usage/batch \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
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
  }'
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "message": "批量创建完成",
    "success_count": 2,
    "failure_count": 0,
    "record_ids": ["uuid1", "uuid2"],
    "errors": []
  }
}
```

#### 3. 查询记录详情

**GET** `/api/v1/risk-report/usage/:id`

请求示例：

```bash
curl -X GET http://localhost:8080/api/v1/risk-report/usage/uuid-here \
  -H "X-API-Key: your-api-key"
```

#### 4. 查询记录列表

**GET** `/api/v1/risk-report/usage`

查询参数：
- `user_id`: 用户 ID（可选）
- `ticker`: 股票代码（可选）
- `start_time`: 开始时间，RFC3339 格式（可选）
- `end_time`: 结束时间，RFC3339 格式（可选）
- `page`: 页码，默认 1
- `page_size`: 每页数量，默认 20

请求示例：

```bash
curl -X GET "http://localhost:8080/api/v1/risk-report/usage?user_id=123456789&page=1&page_size=20" \
  -H "X-API-Key: your-api-key"
```

#### 5. 查询用户统计信息

**GET** `/api/v1/risk-report/usage/stats/:user_id`

查询参数：
- `start_time`: 开始时间，RFC3339 格式（可选）
- `end_time`: 结束时间，RFC3339 格式（可选）

请求示例：

```bash
curl -X GET "http://localhost:8080/api/v1/risk-report/usage/stats/123456789" \
  -H "X-API-Key: your-api-key"
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_queries": 150,
    "total_tokens": 368000,
    "total_prompt_tokens": 280800,
    "total_completion_tokens": 87200,
    "avg_response_time_ms": 6850
  }
}
```

## 配置说明

### 1. API Key 配置

在 `configs/config.yaml` 中配置 API Keys：

```yaml
risk_report:
  api_keys:
    - "risk-report-prod-key-your-secret-here"
    - "risk-report-dev-key-another-key"
```

**重要：** 生产环境务必修改默认的 API Key！

### 2. 环境变量配置

也可以通过环境变量配置（优先级更高）：

```bash
export APP_RISK_REPORT_API_KEYS="key1,key2,key3"
```

## 数据验证规则

### 必填字段

- `user_id`: 用户 ID
- `ticker`: 股票代码（1-10 个大写字母/数字/点号）
- `request_time`: 请求时间（RFC3339 格式）
- `response_time`: 响应时间（RFC3339 格式）
- `prompt_tokens`: Prompt Token 数（>=0）
- `completion_tokens`: Completion Token 数（>=0）
- `total_tokens`: 总 Token 数（应等于 prompt + completion）
- `ai_response`: AI 响应内容

### 验证规则

1. `ticker` 格式：`^[A-Z0-9.]{1,10}$`
2. `request_time` 不能晚于 `response_time`
3. `response_time` 不能是未来时间（允许 5 分钟误差）
4. `total_tokens` = `prompt_tokens` + `completion_tokens`
5. Token 数量必须 >= 0
6. `market_state` 必须是 `PRE`/`REGULAR`/`POST`/`CLOSED` 之一（如果提供）

## 错误码

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| 10002 | 401 | 未提供 API Key 或 API Key 无效 |
| 10001 | 400 | 请求参数错误 |
| 10007 | 400 | 数据验证失败 |
| 10006 | 500 | 服务器内部错误 |

## 测试

项目提供了测试脚本，可以快速验证功能：

```bash
# 确保服务已启动
make run

# 在另一个终端运行测试
./scripts/test_risk_report_api.sh
```

## risk-report 项目集成示例

### Python 客户端示例

```python
import httpx
import asyncio
from datetime import datetime, timezone

class UsageReporter:
    def __init__(self, api_url: str, api_key: str):
        self.api_url = api_url
        self.api_key = api_key
        self.client = httpx.AsyncClient(timeout=5.0)
    
    async def report(self, data: dict):
        """上报使用记录"""
        try:
            response = await self.client.post(
                f"{self.api_url}/api/v1/risk-report/usage",
                json=data,
                headers={"X-API-Key": self.api_key}
            )
            response.raise_for_status()
            return response.json()
        except Exception as e:
            # 失败时只记录日志，不影响主流程
            print(f"上报失败: {e}")
            return None

# 使用示例
async def main():
    reporter = UsageReporter(
        api_url="http://localhost:8080",
        api_key="your-api-key"
    )
    
    data = {
        "user_id": "123456789",
        "ticker": "AAPL",
        "request_time": datetime.now(timezone.utc).isoformat(),
        "response_time": datetime.now(timezone.utc).isoformat(),
        "prompt_tokens": 1872,
        "completion_tokens": 580,
        "total_tokens": 2452,
        "ai_response": "AI 分析结果...",
        "stock_price": 259.83,
        "market_state": "PRE"
    }
    
    result = await reporter.report(data)
    print(f"上报结果: {result}")

if __name__ == "__main__":
    asyncio.run(main())
```

## 数据库表结构

```sql
CREATE TABLE risk_report_usage (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    ticker VARCHAR(10) NOT NULL,
    request_time DATETIME NOT NULL,
    response_time DATETIME NOT NULL,
    prompt_tokens INT NOT NULL,
    completion_tokens INT NOT NULL,
    total_tokens INT NOT NULL,
    ai_response TEXT NOT NULL,
    
    stock_price DECIMAL(10,2),
    market_state VARCHAR(20),
    news_sentiment_score INT,
    news_sentiment_label VARCHAR(20),
    peak_signals_triggered INT,
    action_suggestion VARCHAR(50),
    rate_limit_remaining INT,
    error_message TEXT,
    response_duration_ms INT,
    
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_ticker (ticker),
    INDEX idx_request_time (request_time)
);
```

## 性能优化建议

1. **异步上报**：在 risk-report 项目中使用异步上报，避免阻塞用户响应
2. **失败重试**：上报失败时可以记录到本地队列，定期重试
3. **批量上报**：如果有多条记录，使用批量接口可以提高性能
4. **超时控制**：设置合理的超时时间（建议 5 秒）

## 监控与告警

建议监控以下指标：
- 上报成功率（应 > 95%）
- 平均响应时间（应 < 100ms）
- Token 消耗趋势
- 用户查询频率

## 常见问题

### Q: API Key 忘记了怎么办？
A: 在服务器上查看 `configs/config.yaml` 文件中的 `risk_report.api_keys` 配置。

### Q: 如何添加新的 API Key？
A: 编辑配置文件添加新的 key，然后重启服务。无需重启数据库。

### Q: 上报失败会影响主业务吗？
A: 不会。建议在客户端使用异步上报，失败时只记录日志。

### Q: 数据会保留多久？
A: 目前没有自动清理机制，建议根据需求定期手动归档或清理历史数据。
