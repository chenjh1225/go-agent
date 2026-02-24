# 交互式对话示例 - 功能说明
## 已实现功能
### 1. RAG 上下文记忆 ✓
- 自动存储对话历史到向量数据库
- 语义检索相关历史记忆
- MMR 去重算法
- 重要性评分和时间衰减
### 2. 工具调用 ✓
- **天气工具** - 查询城市天气（温度、湿度、天气状况）
- **时间工具** - 获取当前时间（支持时区、多种格式）
- **计算器工具** - 执行数学运算（加减乘除）
### 3. 私有部署支持 ✓
- 支持通过环境变量设置 \OPENAI_BASE_URL\
- 支持通过命令行参数 \-baseurl\ 设置
- 兼容 OpenAI 协议的私有 API
### 4. 多模型支持 ✓
- OpenAI (GPT-4, GPT-3.5)
- Google Gemini
- Anthropic Claude
- Ollama (本地部署)
## 文件结构
\\\
cmd/example/interactive_chat/
├── main.go       # 主程序（500+ 行）
├── README.md     # 快速开始指南
└── USAGE.md      # 详细使用指南
\\\
## 快速测试
\\\ash
# 1. 设置环境变量
export OPENAI_API_KEY="your-key"
# 2. 运行
cd cmd/example/interactive_chat
go run main.go
# 3. 测试功能
你: 北京今天天气怎么样？  # 测试工具调用
你: 我叫张三              # 测试记忆
你: 我的名字是什么？      # 测试记忆检索
\\\
## 代码修改
### src/models/openai.go
添加了 \NewOpenAILLMWithBaseURL\ 函数，支持自定义 BaseURL
### docs/项目架构文档.md
更新了 OpenAI 模型部分，添加私有部署配置说明
## 下一步建议
1. **接入真实 API** - 将模拟天气数据替换为真实 API
2. **持久化存储** - 使用 PostgreSQL 或 Qdrant 替代内存存储
3. **添加更多工具** - 如搜索、翻译、图片生成等
4. **优化系统提示词** - 根据实际使用调整 AI 行为
