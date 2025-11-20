# README

将要实现一个客户端工具，但从MVP的角度，我们先验证功能的可用性。

这个项目预计是golang实现的，如果你觉得有别的方案，可以在golang，python中选择。

我们先实现一个动态配置更新的server，这个server将提供和openai一样的endpoint，并生成一个uri作为api_base 以接收请求。

程序既是一个cli，又是一个server。
用户可以启动cli
- 通过命令行交互
  - 新增，配置name，api-base，和对应的token，这些机密信息需要妥善保存在固定的位置
  - 查看，
  - 删除
- 通过命令行交互，查看和管理服务状态，你需要选择合适的方案管理服务


服务提供的能力
- 提供统一的uri api-base，可以由任意client配置使用
- 生成token，供client 使用
- 加载配置，和刷新配置
  - 注意，当cli对配置进行了调整时，需要即时进行刷新
- client使用token，并选中模型后，会链接到server，此时server需要对应转换字段到正确的api-base和token，完成真正的大模型请求