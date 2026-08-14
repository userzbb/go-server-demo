# 贡献指南

在参与开发前，请阅读并遵守以下规范。

## 开发规范
所有开发活动必须遵守 [docs/STANDARDS.md](docs/STANDARDS.md)。

## 提交前自检
1. 运行 `make ci` 确保本地检查通过
2. 确保新增代码有对应的单元测试
3. 确保提交信息符合规范（`feat:` / `fix:` / `docs:` / `chore:`）

## Git 工作流
1. 从 `develop` 分支创建功能分支：`git checkout -b feature/xxx`
2. 开发完成后推送到远程仓库
3. 提交 Pull Request 合并到 `develop`
