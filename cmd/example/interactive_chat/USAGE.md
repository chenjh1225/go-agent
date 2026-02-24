# 交互式对话示例 - 详细使用指南
## 概述
这是一个功能完整的交互式 AI 对话示例，展示了 Go-Agent 框架的核心能力：
1. **RAG 上下文记忆** - 自动记住并检索相关对话历史
2. **工具调用** - 智能调用外部工具获取实时信息
3. **多模型支持** - 支持多种 LLM 提供商
4. **私有部署** - 支持 OpenAI 兼容的私有 API
## 功能特性
### 1. RAG 上下文记忆
程序会自动：
- 将每轮对话存储到向量数据库
- 在回答前检索相关历史对话
- 使用语义相似度排序记忆
- 支持 MMR 去重避免重复信息
**示例**：
\\\
你: 我叫张三
AI: 你好，张三！
你: 我的名字是什么？
AI: 你的名字是张三。（从记忆中检索）
\\\
### 2. 工具调用
内置三个实用工具：
#### 天气工具
- 名称：\get_weather\
- 功能：查询城市天气
- 参数：城市名、温度单位
#### 时间工具
- 名称：\get_current_time\
- 功能：获取当前时间
- 参数：时区、格式
#### 计算器工具
- 名称：\calculator\
- 功能：数学运算
- 参数：操作类型、两个操作数
## 安装与配置
### 环境要求
- Go 1.18 或更高版本
- 网络连接（用于调用 LLM API）
### 配置步骤
#### 1. 使用 OpenAI
\\\ash
# 官方 API
export OPENAI_API_KEY="sk-your-api-key"
# 私有部署（兼容 OpenAI 协议）
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://your-private-api.com/v1"
\\\
#### 2. 使用 Google Gemini
\\\ash
export GOOGLE_API_KEY="your-google-api-key"
\\\
#### 3. 使用 Anthropic Claude
\\\ash
export ANTHROPIC_API_KEY="your-anthropic-key"
\\\
#### 4. 使用 Ollama（本地）
\\\ash
# 启动 Ollama 服务
ollama serve
# 拉取模型
ollama pull llama2
\\\
## 运行示例
### 基本用法
\\\ash
cd F:\study\go-agent\cmd\example\interactive_chat
go run main.go
\\\
### 高级用法
\\\ash
# 指定提供商和模型
go run main.go -provider openai -model gpt-4
# 使用私有部署
go run main.go -baseurl "https://your-api.com/v1"
# 使用 Gemini
go run main.go -provider gemini -model gemini-2.5-pro
# 使用本地模型
go run main.go -provider ollama -model llama2
# 指定会话 ID（恢复历史对话）
go run main.go -session my-chat-123
\\\
## 使用示例
### 场景 1: 天气查询
\\\
你: 北京今天天气怎么样？
AI: 让我查询一下北京的天气...
[工具调用: get_weather]
城市：北京
天气状况：晴朗
温度：25°C
湿度：60%
风力：微风
空气质量：良好
\\\
### 场景 2: 时间查询
\\\
你: 现在几点了？
AI: [工具调用: get_current_time]
当前时间：2026年02月24日 15:30:00 星期一
时区：Asia/Shanghai
\\\
### 场景 3: 数学计算
\\\
你: 帮我计算 123 * 456
AI: [工具调用: calculator]
123.00 × 456.00 = 56088.00
\\\
### 场景 4: 记忆测试
\\\
你: 记住，我的生日是 5 月 20 日
AI: 好的，我记住了你的生日是 5 月 20 日。
（聊其他内容...）
你: 我的生日是哪天？
AI: 根据我们之前的对话，你的生日是 5 月 20 日。
\\\
### 场景 5: 复杂查询
\\\
你: 如果北京现在是下午 3 点，纽约现在几点？
AI: [工具调用: get_current_time，查询两地时间]
北京现在是 15:00，纽约现在是 02:00（考虑13小时时差）。
\\\
## 交互式命令
| 命令 | 功能 |
|------|------|
| \exit\ 或 \quit\ | 退出程序 |
| \clear\ | 清空当前会话记忆 |
| \help\ | 显示帮助信息 |
## 命令行参数
| 参数 | 默认值 | 说明 |
|------|--------|------|
| \-provider\ | \openai\ | LLM 提供商 |
| \-model\ | \gpt-4o-mini\ | 模型名称 |
| \-session\ | 自动生成 | 会话 ID |
| \-baseurl\ | 空 | API 基础 URL（私有部署） |
## 技术实现
### RAG 记忆流程
1. **存储**：用户输入和 AI 响应 → 生成向量 → 存入向量数据库
2. **检索**：新用户输入 → 生成查询向量 → 检索相似记忆
3. **整合**：相关记忆 + 新输入 → 发送给 LLM
4. **生成**：LLM 基于上下文生成回答
### 工具调用流程
1. **识别**：LLM 判断是否需要工具
2. **提取**：从响应中提取工具名和参数
3. **执行**：调用对应工具的 \Invoke\ 方法
4. **整合**：将工具结果返回给 LLM
5. **生成**：LLM 基于工具结果生成最终回答
## 扩展开发
### 添加自定义工具
\\\go
// 1. 定义工具结构
type MyTool struct{}
// 2. 实现 Spec 方法
func (t *MyTool) Spec() agent.ToolSpec {
    return agent.ToolSpec{
        Name:        "my_tool",
        Description: "工具描述",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "param": map[string]any{
                    "type": "string",
                    "description": "参数描述",
                },
            },
        },
    }
}
// 3. 实现 Invoke 方法
func (t *MyTool) Invoke(ctx context.Context, req agent.ToolRequest) (agent.ToolResponse, error) {
    // 工具逻辑
    return agent.ToolResponse{
        Content: "结果",
    }, nil
}
// 4. 在 main.go 中注册
tools := []agent.Tool{
    NewMyTool(),
    // ...其他工具
}
\\\
### 接入真实天气 API
修改 \simulateWeatherQuery\ 函数：
\\\go
func fetchRealWeather(city string) (string, error) {
    // 调用真实 API（如 OpenWeatherMap）
    apiKey := os.Getenv("WEATHER_API_KEY")
    url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", city, apiKey)
    resp, err := http.Get(url)
    // 处理响应...
}
\\\
## 故障排除
### 问题 1: API Key 未设置
**错误**：\OPENAI_API_KEY 环境变量未设置\
**解决**：
\\\ash
export OPENAI_API_KEY="your-key"
\\\
### 问题 2: 私有部署连接失败
**错误**：\connection refused\ 或 \	imeout\
**解决**：
1. 检查 BaseURL 格式（通常以 \/v1\ 结尾）
2. 测试连接：\curl https://your-api.com/v1/models\
3. 检查防火墙设置
### 问题 3: 工具未被调用
**原因**：某些模型不支持工具调用或提示不够明确
**解决**：
1. 使用支持工具调用的模型（GPT-4, Gemini Pro）
2. 更明确的提示："使用天气工具查询北京天气"
### 问题 4: 记忆功能异常
**原因**：嵌入模型未配置
**解决**：
\\\ash
# 设置嵌入提供商
export ADK_EMBED_PROVIDER="openai"
export OPENAI_API_KEY="your-key"
\\\
## 性能优化建议
1. **使用持久化存储**：将 InMemory 改为 PostgreSQL/Qdrant
2. **批量嵌入**：减少 API 调用次数
3. **缓存工具规范**：避免重复序列化
4. **并发工具调用**：多个工具同时执行
## 相关资源
- [项目架构文档](../../../docs/项目架构文档.md)
- [Go-Agent GitHub](https://github.com/Protocol-Lattice/go-agent)
- [OpenAI API 文档](https://platform.openai.com/docs)
- [Gemini API 文档](https://ai.google.dev/docs)
## 许可证
MIT License
