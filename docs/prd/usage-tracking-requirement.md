# 使用记录上报需求文档

## 业务背景
risk-report 项目需要记录每次用户查询的详细信息，用于分析、监控和成本核算。数据存储在独立的 go-user-api 服务中，通过 HTTP API 上报。

## 数据字段

### 核心字段（必填）
| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| `user_id` | string/int64 | Telegram 用户 ID | `123456789` |
| `ticker` | string | 查询的股票代码 | `AAPL` |
| `request_time` | timestamp/string | 用户发起查询的时间 | `2026-01-15T10:30:45Z` |
| `response_time` | timestamp/string | 机器人返回结果的时间 | `2026-01-15T10:30:52Z` |
| `prompt_tokens` | int | AI 提示词消耗的 token 数 | `1200` |
| `completion_tokens` | int | AI 生成回复消耗的 token 数 | `450` |
| `total_tokens` | int | 总消耗 token 数 | `1650` |
| `ai_response` | text | AI 分析的原文（完整回复） | `【当前阶段判断】...` |

### 扩展字段（可选）
| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| `stock_price` | float | 查询时的股票价格 | `259.83` |
| `market_state` | string | 市场状态 | `PRE/REGULAR/POST/CLOSED` |
| `news_sentiment_score` | int | 新闻情绪分数 | `90` |
| `news_sentiment_label` | string | 新闻情绪标签 | `偏多/中性/偏空` |
| `peak_signals_triggered` | int | 触发的见顶信号数 | `2` |
| `action_suggestion` | string | 操作建议 | `偏买入/试探` |
| `rate_limit_remaining` | int | 用户剩余查询额度 | `8` |
| `error_message` | string | 错误信息（如查询失败） | `未找到该股票代码` |
| `response_duration_ms` | int | 响应耗时（毫秒） | `7200` |

---

## API 接口设计建议

### 接口 1：上报查询记录（推荐）

**请求**
```http
POST /api/v1/risk-report/usage
Content-Type: application/json
X-API-Key: <your-secret-key>

{
  "user_id": "123456789",
  "ticker": "AAPL",
  "request_time": "2026-01-15T10:30:45Z",
  "response_time": "2026-01-15T10:30:52Z",
  "prompt_tokens": 1200,
  "completion_tokens": 450,
  "total_tokens": 1650,
  "ai_response": "【当前阶段判断】...",
  "stock_price": 259.83,
  "market_state": "PRE",
  "news_sentiment_score": 90,
  "news_sentiment_label": "偏多",
  "peak_signals_triggered": 2,
  "action_suggestion": "偏买入/试探",
  "rate_limit_remaining": 8,
  "response_duration_ms": 7200
}
```

**响应**
```json
{
  "success": true,
  "message": "记录已保存",
  "record_id": "abc123"
}
```

**错误响应**
```json
{
  "success": false,
  "error": "invalid_api_key",
  "message": "API Key 无效"
}
```

### 接口 2：批量上报（可选，如有性能需求）

**请求**
```http
POST /api/v1/risk-report/usage/batch
Content-Type: application/json
X-API-Key: <your-secret-key>

{
  "records": [
    { /* 记录1 */ },
    { /* 记录2 */ }
  ]
}
```

---

## go-user-api 实现要点

### 1. 数据库表设计（参考）

```sql
CREATE TABLE risk_report_usage (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    ticker VARCHAR(10) NOT NULL,
    request_time TIMESTAMP NOT NULL,
    response_time TIMESTAMP NOT NULL,
    prompt_tokens INT NOT NULL,
    completion_tokens INT NOT NULL,
    total_tokens INT NOT NULL,
    ai_response TEXT NOT NULL,
    
    -- 扩展字段
    stock_price DECIMAL(10,2),
    market_state VARCHAR(20),
    news_sentiment_score INT,
    news_sentiment_label VARCHAR(20),
    peak_signals_triggered INT,
    action_suggestion VARCHAR(50),
    rate_limit_remaining INT,
    error_message TEXT,
    response_duration_ms INT,
    
    -- 系统字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_ticker (ticker),
    INDEX idx_request_time (request_time)
);
```

### 2. API Key 认证
- 从请求头 `X-API-Key` 读取
- 验证 key 是否有效
- 建议支持多个 key（为不同服务分配）

### 3. 数据校验
- `user_id`、`ticker`、`request_time`、`response_time`、token 字段必填
- `ticker` 格式：1-6 个大写字母/数字/点号
- `request_time` <= `response_time`
- token 数值 >= 0

### 4. 性能优化（可选）
- 异步写入队列（如 Redis/消息队列）
- 批量插入数据库
- 定期归档历史数据

---

## risk-report 项目实现方案

### 在哪里埋点上报

**关键位置：** `src/handlers/message_handler.py` 的 `handle_ticker()` 方法

```python
async def handle_ticker(self, message: Message):
    ticker = message.text.strip().upper()
    user_id = message.from_user.id
    request_time = datetime.now(timezone.utc)  # 记录请求时间
    
    # ... 现有的验证逻辑 ...
    
    # 执行分析
    result = await self.analysis_service.analyze_stock(ticker)
    response_time = datetime.now(timezone.utc)  # 记录响应时间
    
    # 上报使用记录
    await self._report_usage(
        user_id=user_id,
        ticker=ticker,
        request_time=request_time,
        response_time=response_time,
        result=result,
        used=used,
        limit=limit
    )
    
    # ... 现有的回复逻辑 ...
```

### 实现细节

1. **新增配置项**（`.env`）
```bash
# 使用记录上报配置
USAGE_REPORT_ENABLED=true
USAGE_REPORT_URL=http://your-go-api-domain/api/v1/risk-report/usage
USAGE_REPORT_API_KEY=your-secret-key-here
USAGE_REPORT_TIMEOUT=5
```

2. **新增上报客户端**（`src/api/usage_reporter.py`）
```python
class UsageReporter:
    async def report(self, data: dict):
        # HTTP POST 上报
        # 失败重试
        # 超时控制
```

3. **集成到 MessageHandler**
- 在 `handle_ticker` 成功/失败后都上报
- 异步上报，不阻塞用户响应
- 上报失败只记录日志，不影响主流程

4. **提取 token 消耗**
- DeepSeek API 响应包含 `usage` 字段
- 需要在 `DeepSeekClient.chat_completion()` 返回 token 信息

---

## 配置示例

### go-user-api 侧
```yaml
# config.yaml
api:
  keys:
    - "risk-report-prod-key-xxx"
    - "risk-report-dev-key-yyy"

database:
  host: localhost
  port: 5432
  name: user_db
```

### risk-report 侧
```bash
# .env
USAGE_REPORT_ENABLED=true
USAGE_REPORT_URL=http://localhost:8080/api/v1/risk-report/usage
USAGE_REPORT_API_KEY=risk-report-prod-key-xxx
USAGE_REPORT_TIMEOUT=5
```

---

## 时间线建议

1. **go-user-api 开发**（1-2 天）
   - 设计并创建数据库表
   - 实现 POST `/api/v1/risk-report/usage` 接口
   - API Key 认证
   - 单元测试

2. **risk-report 集成**（半天）
   - 新增 `UsageReporter` 客户端
   - 在 `MessageHandler` 埋点上报
   - 配置项与测试

3. **联调测试**（半天）
   - 本地环境验证
   - 查看数据库记录是否正确
   - 压力测试（可选）

---

## 注意事项

1. **隐私合规**：确保用户知晓数据收集，符合相关法律（GDPR/隐私政策）
2. **数据脱敏**：如需分析，考虑对 `user_id` 做哈希处理
3. **日志分离**：上报失败不应影响主业务，记录到独立日志
4. **监控告警**：上报成功率低于 95% 时发送告警
5. **数据归档**：定期归档 30 天以上的历史数据

---

## 示例：完整的上报数据

```json
{
  "user_id": "123456789",
  "ticker": "AAPL",
  "request_time": "2026-01-15T02:30:45.123Z",
  "response_time": "2026-01-15T02:30:52.456Z",
  "prompt_tokens": 1872,
  "completion_tokens": 580,
  "total_tokens": 2452,
  "ai_response": "【当前阶段判断】★★★★★\n明确判断：高位回调\n主要依据：距高点-10%+RSI超卖25.6+成交萎缩=回调中期\n所处位置：52周区间75.9%位（偏高），RSI 25.6（超卖）\n\n【见顶信号检查】★★★★\n量价背离：否\n技术指标：RSI 25.6未超买，MA50在上MA200多头排列\n极端位置：75.9%未到极端区，距高点-10%非顶部\n情绪狂热：新闻情绪偏多但未极端\n→ 综合判断：已触发1/4个顶部信号，见顶概率20-35%\n\n【短期走势预判】（1-4周）\n方向+概率：震荡反弹（60%概率）\n核心逻辑：RSI超卖叠加支撑位233.5附近，大概率技术反弹；但高位震荡格局未改，反弹空间有限\n\n【操作建议】★★★★★\n建议动作：买入/轻仓试探\n具体策略：\n  若买入：价位$259-261（当前价or小幅回调），目标$272（MA50阻力），止损$245（-5.7%，支撑下破）\n  若卖出：反弹至$272附近减仓，目标持有现金\n  若观望：等待跌破$245止损或突破$272追涨信号\n关键支撑：$233.5（MA200）\n关键阻力：$272.2（MA50）\n\n【风险提示】\n利空（2点）：\n1. 估值仍处高位（PE不详），整体科技板块调整风险\n2. 成交量萎缩显示买盘不强，反弹持续性存疑\n利好（2点）：\n1. RSI超卖+距MA200支撑近，技术面反弹动力充足\n2. 新闻情绪偏多（beat/surge关键词），短期情绪支撑\n\n【一句话总结】\n高位技术性回调后超卖反弹，可轻仓试探259-261区间，目标272，严守245止损",
  "stock_price": 259.83,
  "market_state": "PRE",
  "news_sentiment_score": 90,
  "news_sentiment_label": "偏多",
  "peak_signals_triggered": 1,
  "action_suggestion": "偏买入/试探",
  "rate_limit_remaining": 8,
  "response_duration_ms": 7311
}
```
