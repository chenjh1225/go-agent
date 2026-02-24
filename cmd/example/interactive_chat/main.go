// main.go - 交互式对话示例
// 功能：
//   1. 使用 RAG 实现上下文记忆
//   2. 集成工具调用（天气查询等）
//   3. 支持多种 LLM 提供商
//
// 使用方法：
//   export OPENAI_API_KEY=your-api-key
//   export OPENAI_BASE_URL=https://your-private-deployment.com/v1  # 可选，用于私有部署
//   go run main.go
//
// 或者使用其他提供商：
//   export GOOGLE_API_KEY=your-api-key
//   go run main.go -provider gemini -model gemini-2.5-pro

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Protocol-Lattice/go-agent"
	"github.com/Protocol-Lattice/go-agent/src/adk"
	"github.com/Protocol-Lattice/go-agent/src/adk/modules"
	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/models"
)

var (
	flagProvider = flag.String("provider", "openai", "LLM 提供商: openai|gemini|anthropic|ollama")
	flagModel    = flag.String("model", "gpt-4o-mini", "模型名称")
	flagSession  = flag.String("session", "", "会话 ID（留空则自动生成）")
	flagBaseURL  = flag.String("baseurl", "", "自定义 API 基础 URL（用于私有部署）")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	// 生成或使用指定的会话 ID
	sessionID := *flagSession
	if sessionID == "" {
		sessionID = fmt.Sprintf("chat-%d", time.Now().Unix())
	}

	fmt.Println("=== Go-Agent 交互式对话示例 ===")
	fmt.Printf("提供商: %s\n", *flagProvider)
	fmt.Printf("模型: %s\n", *flagModel)
	fmt.Printf("会话 ID: %s\n", sessionID)
	if *flagBaseURL != "" {
		fmt.Printf("API URL: %s\n", *flagBaseURL)
	}
	fmt.Println("\n功能特性:")
	fmt.Println("✓ RAG 上下文记忆 - 自动记住对话历史")
	fmt.Println("✓ 工具调用 - 可查询天气、时间等信息")
	fmt.Println("✓ 语义检索 - 智能关联相关历史对话")
	fmt.Println("\n输入 'exit' 或 'quit' 退出，输入 'clear' 清空记忆\n")

	// 创建 Agent
	ag, err := createAgent(ctx, *flagProvider, *flagModel, *flagBaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建 Agent 失败: %v\n", err)
		os.Exit(1)
	}

	// 交互式对话循环
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("你: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// 处理特殊命令
		switch strings.ToLower(input) {
		case "exit", "quit":
			fmt.Println("再见！")
			return
		case "clear":
			// 切换到新会话 ID 来清空记忆
			sessionID = fmt.Sprintf("chat-%d", time.Now().Unix())
			fmt.Printf("✓ 已清空记忆，新会话 ID: %s\n", sessionID)
			continue
		case "help":
			printHelp()
			continue
		}

		// 生成响应
		fmt.Print("AI: ")
		response, err := ag.Generate(ctx, sessionID, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 错误: %v\n", err)
			continue
		}

		fmt.Println(response)
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "读取输入错误: %v\n", err)
	}
}

// createAgent 创建配置好的 Agent 实例
func createAgent(ctx context.Context, provider, model, baseURL string) (*agent.Agent, error) {
	// 系统提示词
	systemPrompt := `你是一个友好的 AI 助手。你拥有以下能力：

1. **上下文记忆**：你能记住对话历史，并在回答时考虑之前的对话内容。
2. **工具调用**：当用户询问天气、时间等信息时，你会自动调用相应的工具获取实时数据。
3. **智能理解**：你能理解用户的意图，并提供准确、有帮助的回答。

注意事项：
- 当用户询问天气时，使用 get_weather 工具
- 当用户询问时间时，使用 get_current_time 工具
- 保持回答简洁、友好、准确`

	// 创建 ADK Kit
	kit, err := adk.New(ctx,
		adk.WithDefaultSystemPrompt(systemPrompt),
		adk.WithModules(
			// 模型模块
			modules.NewModelModule(provider, func(ctx context.Context) (models.Agent, error) {
				return createModel(provider, model, baseURL, systemPrompt)
			}),
			// 内存模块 - 使用 InMemory 存储，支持 RAG
			modules.InMemoryMemory(
				100000,                // 上下文限制
				memory.AutoEmbedder(), // 自动选择嵌入模型
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 ADK Kit 失败: %w", err)
	}

	// 注册工具
	tools := []agent.Tool{
		NewWeatherTool(),
		NewTimeTool(),
		NewCalculatorTool(),
	}

	for _, tool := range tools {
		if err := kit.RegisterTool(tool); err != nil {
			return nil, fmt.Errorf("注册工具失败: %w", err)
		}
	}

	// 构建 Agent
	ag, err := kit.BuildAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("构建 Agent 失败: %w", err)
	}

	return ag, nil
}

// createModel 根据提供商创建模型实例
func createModel(provider, model, baseURL, systemPrompt string) (models.Agent, error) {
	switch strings.ToLower(provider) {
	case "openai":
		if baseURL != "" {
			return models.NewOpenAILLMWithBaseURL(model, systemPrompt, baseURL)
		}
		return models.NewOpenAILLM(model, systemPrompt), nil
	case "gemini":
		return models.NewGeminiLLM(context.Background(), model, systemPrompt)
	case "anthropic":
		return models.NewAnthropicLLM(model, systemPrompt)
	case "ollama":
		return models.NewOllamaLLM(model, systemPrompt)
	default:
		return nil, fmt.Errorf("不支持的提供商: %s", provider)
	}
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("\n=== 可用命令 ===")
	fmt.Println("exit/quit  - 退出程序")
	fmt.Println("clear      - 清空对话记忆")
	fmt.Println("help       - 显示此帮助信息")
	fmt.Println("\n=== 示例问题 ===")
	fmt.Println("• 北京今天天气怎么样？")
	fmt.Println("• 现在几点了？")
	fmt.Println("• 计算 123 * 456")
	fmt.Println("• 我之前问过什么问题？（测试记忆功能）")
	fmt.Println()
}

// ============================================================================
// 工具实现
// ============================================================================

// WeatherTool - 天气查询工具
type WeatherTool struct{}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (t *WeatherTool) Spec() agent.ToolSpec {
	return agent.ToolSpec{
		Name:        "get_weather",
		Description: "获取指定城市的当前天气信息，包括温度、天气状况、湿度等",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"city": map[string]any{
					"type":        "string",
					"description": "城市名称，如：北京、上海、New York",
				},
				"unit": map[string]any{
					"type":        "string",
					"description": "温度单位，celsius（摄氏度）或 fahrenheit（华氏度），默认为 celsius",
					"enum":        []string{"celsius", "fahrenheit"},
				},
			},
			"required": []string{"city"},
		},
	}
}

func (t *WeatherTool) Invoke(ctx context.Context, req agent.ToolRequest) (agent.ToolResponse, error) {
	// 解析参数
	args := req.Arguments
	city, ok := args["city"].(string)
	if !ok || city == "" {
		return agent.ToolResponse{
			Error: "缺少必需参数: city",
		}, nil
	}

	unit := "celsius"
	if u, ok := args["unit"].(string); ok {
		unit = u
	}

	// 模拟天气查询（实际应用中应调用真实的天气 API）
	weather := simulateWeatherQuery(city, unit)

	return agent.ToolResponse{
		Content: weather,
	}, nil
}

// simulateWeatherQuery 模拟天气查询（演示用）
func simulateWeatherQuery(city, unit string) string {
	// 在实际应用中，这里应该调用真实的天气 API
	// 例如：OpenWeatherMap, WeatherAPI 等

	rand.Seed(time.Now().UnixNano())
	temp := 15 + rand.Intn(20) // 15-35度

	if unit == "fahrenheit" {
		temp = int(float64(temp)*1.8 + 32)
	}

	conditions := []string{"晴朗", "多云", "阴天", "小雨", "晴转多云"}
	condition := conditions[rand.Intn(len(conditions))]

	humidity := 40 + rand.Intn(40) // 40-80%

	unitStr := "°C"
	if unit == "fahrenheit" {
		unitStr = "°F"
	}

	result := fmt.Sprintf(`城市：%s
天气状况：%s
温度：%d%s
湿度：%d%%
风力：微风
空气质量：良好

注意：这是模拟数据。实际应用中应接入真实的天气 API。`,
		city, condition, temp, unitStr, humidity)

	return result
}

// TimeTool - 时间查询工具
type TimeTool struct{}

func NewTimeTool() *TimeTool {
	return &TimeTool{}
}

func (t *TimeTool) Spec() agent.ToolSpec {
	return agent.ToolSpec{
		Name:        "get_current_time",
		Description: "获取当前日期和时间",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"timezone": map[string]any{
					"type":        "string",
					"description": "时区，如：Asia/Shanghai, America/New_York，留空则使用本地时区",
				},
				"format": map[string]any{
					"type":        "string",
					"description": "时间格式，如：full（完整）, date（仅日期）, time（仅时间）",
					"enum":        []string{"full", "date", "time"},
				},
			},
		},
	}
}

func (t *TimeTool) Invoke(ctx context.Context, req agent.ToolRequest) (agent.ToolResponse, error) {
	args := req.Arguments

	// 获取时区
	var loc *time.Location
	var err error
	if tz, ok := args["timezone"].(string); ok && tz != "" {
		loc, err = time.LoadLocation(tz)
		if err != nil {
			return agent.ToolResponse{
				Error: fmt.Sprintf("无效的时区: %s", tz),
			}, nil
		}
	} else {
		loc = time.Local
	}

	now := time.Now().In(loc)

	// 格式化输出
	format := "full"
	if f, ok := args["format"].(string); ok {
		format = f
	}

	var result string
	switch format {
	case "date":
		result = fmt.Sprintf("日期：%s", now.Format("2006年01月02日 星期一"))
	case "time":
		result = fmt.Sprintf("时间：%s", now.Format("15:04:05"))
	default:
		result = fmt.Sprintf("当前时间：%s\n时区：%s",
			now.Format("2006年01月02日 15:04:05 星期一"),
			loc.String())
	}

	return agent.ToolResponse{
		Content: result,
	}, nil
}

// CalculatorTool - 计算器工具
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Spec() agent.ToolSpec {
	return agent.ToolSpec{
		Name:        "calculator",
		Description: "执行基本的数学运算（加、减、乘、除）",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"operation": map[string]any{
					"type":        "string",
					"description": "运算类型：add（加）, subtract（减）, multiply（乘）, divide（除）",
					"enum":        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": map[string]any{
					"type":        "number",
					"description": "第一个操作数",
				},
				"b": map[string]any{
					"type":        "number",
					"description": "第二个操作数",
				},
			},
			"required": []string{"operation", "a", "b"},
		},
	}
}

func (t *CalculatorTool) Invoke(ctx context.Context, req agent.ToolRequest) (agent.ToolResponse, error) {
	args := req.Arguments

	// 解析参数
	operation, ok := args["operation"].(string)
	if !ok {
		return agent.ToolResponse{Error: "缺少 operation 参数"}, nil
	}

	// 处理不同类型的数字参数
	var a, b float64

	switch v := args["a"].(type) {
	case float64:
		a = v
	case int:
		a = float64(v)
	case json.Number:
		a, _ = v.Float64()
	default:
		return agent.ToolResponse{Error: "参数 a 必须是数字"}, nil
	}

	switch v := args["b"].(type) {
	case float64:
		b = v
	case int:
		b = float64(v)
	case json.Number:
		b, _ = v.Float64()
	default:
		return agent.ToolResponse{Error: "参数 b 必须是数字"}, nil
	}

	// 执行运算
	var result float64
	var opStr string

	switch operation {
	case "add":
		result = a + b
		opStr = "+"
	case "subtract":
		result = a - b
		opStr = "-"
	case "multiply":
		result = a * b
		opStr = "×"
	case "divide":
		if b == 0 {
			return agent.ToolResponse{Error: "除数不能为零"}, nil
		}
		result = a / b
		opStr = "÷"
	default:
		return agent.ToolResponse{Error: "不支持的运算类型: " + operation}, nil
	}

	output := fmt.Sprintf("%.2f %s %.2f = %.2f", a, opStr, b, result)

	return agent.ToolResponse{
		Content: output,
	}, nil
}
