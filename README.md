<div align="center">
  <h1>Diabetes Agent</h1>
  <div>
    <p>基于 Gin 开发的糖尿病智能体平台服务端</p>
  </div>
</div>

## 介绍
毕设作品，主要面向患者侧，包括 Agent 对话、知识库、血糖记录、运动记录、健康周报等核心功能。本仓库为服务端实现，另有[客户端](https://github.com/zikifaker/diabetes-agent-client)和[MCP服务](https://github.com/zikifaker/diabetes-agent-mcp-server)。

## feat
- [x] Agent 对话
  - [x] ReAct Agent
    - [x] 推送 ReAct 思考过程
    - [x] 推送工具调用结果
    - [x] 上下文压缩(LLM 生成摘要)
  - [x] Agent 配置(最大迭代次数/MCP 工具)
  - [x] 选择模型
  - [x] 上传聊天文件(PNG/JPG/JEPG/GIF/WEBP/Word/PDF/Excel/txt/Markdown)
  - [x] 向量检索知识库
  - [x] 语音输入
- [x] Agent 会话
  - [x] 创建会话
  - [x] 获取会话消息
  - [x] 删除会话
  - [x] 更新会话标题
- [x] 知识库
  - [x] 上传文件(PDF/txt/Markdown)
  - [x] ETL(下载 OSS 文件->语法结构/固定长度切分->向量化存储)
  - [x] 删除文件
  - [x] 下载文件
  - [x] 查询文件(按文件名)
- [x] 血糖记录
  - [x] 增加记录
  - [x] 时间范围查询
- [x] 运动记录
  - [x] 增加记录
  - [x] 删除记录
  - [x] 时间范围查询
- [x] 健康档案
  - [x] 创建
  - [x] 更新
- [x] 健康周报
  - [x] 预览
  - [x] 下载
  - [x] 选择是否开启邮件通知
- [x] 系统消息
  - [x] 分页查询
  - [x] 标记已读
  - [x] 删除
  - [x] 查询未读消息数量
- [x] 登录
  - [x] 注册
  - [x] 密码登录
  - [x] 邮箱验证码登录
