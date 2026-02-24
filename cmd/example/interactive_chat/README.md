# 交互式对话示例
## 功能
- RAG 上下文记忆
- 工具调用（天气、时间、计算器）
- 多模型支持
- 私有部署支持
## 快速开始
`ash
# 设置环境变量
export OPENAI_API_KEY="your-key"
export OPENAI_BASE_URL="https://your-api.com/v1"  # 可选
# 运行
cd cmd/example/interactive_chat
go run main.go
# 使用其他模型
go run main.go -provider gemini -model gemini-2.5-pro
`
## 命令
- exit/quit - 退出
- clear - 清空记忆
- help - 帮助
