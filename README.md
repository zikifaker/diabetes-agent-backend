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
  - [x] Agent 配置(最大迭代次数/MCP工具)
  - [x] Agent 模型切换
  - [x] 上传聊天文件(PNG/JPG/JEPG/GIF/WEBP/Word/PDF/Excel/txt/Markdown)
  - [x] 检索知识库
  - [x] 语音输入
- [x] Agent 会话
  - [x] 创建
  - [x] 获取消息
  - [x] 删除
  - [x] 更新标题
- [x] 知识库
  - [x] 上传文件(PDF/txt/Markdown)
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
- [x] 系统消息
  - [x] 分页查询
  - [x] 标记已读
  - [x] 删除
  - [x] 查询未读消息数量
- [x] 登录
  - [x] 注册
  - [x] 密码登录
  - [x] 验证码登录
