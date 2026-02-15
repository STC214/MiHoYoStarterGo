---
slug: posts/2026/README
title: "项目技术总结报告：MiHoYoStarterGo"
description: "标题上写着呢"
date: 2026-02-14
author: "STC"
tags: ["NOTE"]
featured: false
editable: true
---

---

# <span style="color:red">免责声明：工具所有数据仅你本机存储，不含任何网络数据传输功能（毕竟连自动检测更新都没做），若遇到需要备份的情况请将工具目录整体打包后放在你认为丢不了的地方。恢复的时候只需要解压到非中文和特殊符号目录下即可。其次，工具不含有任何外挂行为，不含有任何内存操作。最后就是，没玩过国际服版本，不知道有什么区别，所以短期内不会做国际服的适配。</span>


# 项目技术总结报告：MiHoYoStarterGo (开发者存档版)

## 项目可执行文件刚需 PaddleOCR-Go。如需使用请于下方介绍中下载对应版本。目前项目使用 1.4.1 版。
## 功能目前只做了原神部分，其他部分待完善。

## 1. 核心 OCR 技术选型

本项目的图像识别能力建立在开源社区优秀的离线推理库之上。

* **技术仓库**：[otiai10/gosseract](https://github.com/otiai10/gosseract) (基于 Tesseract) 或 **[PaddleOCR-Go](https://github.com/otiai10/gosseract)** (推荐参考)。
* *注：根据我们目前的实现，主要依赖于本地编译的 PaddleOCR 推理引擎，确保了对中文（简/繁）极高的识别准确率。*


* **仓库简介**：PaddleOCR 是百度开源的超轻量级 OCR 系统，支持 80 多种语言，其特点是模型小（仅几 M）、速度快、精度高，非常适合嵌入到像 Wails 这样的桌面应用中进行离线推断。

---

## 2. 离线 OCR 的技术优势

* **隐私与安全**：所有截图处理均在本地 `temp_*.png` 文件中完成，并在协程结束时由 `os.Remove` 彻底清理。
* **零成本运行**：无需申请云端 API Key，没有调用次数限制，且在无网络环境下（如飞机、高铁）依然能正常工作。
* **环境干扰排除**：开发过程中发现，若使用小米澎湃系统，需关闭**“晕车缓解功能”**，否则系统级的视觉补偿会干扰截图与坐标定位。

---

## 3. 已实现的核心功能闭环

| 模块 | 技术重点 | 关键函数 / 源代码 |
| --- | --- | --- |
| **自动化监控** | 异步循环检测、窗口置顶、截图识别 | `StartAutomationMonitor` |
| **交互控制** | 支持全局暂停、手动停止监控并回收资源 | `TogglePause`, `StopMonitor` |
| **账号管理** | 密码加密存储、Token 提取、账号删除与同步 | `AddAccount`, `DeleteAccount` |
| **前端 UI** | Vue3 响应式布局、多主题切换、实时监控弹窗 | `App.vue` |

---

## 4. 项目部署与权限声明

为了保证离线 OCR 和注册表读取的权限，打包时必须包含以下设置：

1. **管理员清单**：`build/windows/wails.exe.manifest` 中已配置 `requireAdministrator` 以获取系统句柄。
2. **打包命令**：
```bash
wails build -clean -platform windows/amd64 -ldflags "-s -w"

```

3. **依赖说明**：如果更换开发环境，请确保 `gosseract` 依赖的 `tesseract-ocr` 库或 `Paddle` 推理库已正确配置于系统环境变量中。

---

## 5. 项目总结

从最初的 Go 后端逻辑设计，到 Vue3 前端 UI 的打磨，再到**离线 PaddleOCR** 的深度集成与 **Windows UAC 权限** 的攻克，**MiHoYoStarterGo** 现在已成为一个安全、高效、且具备工业级稳定性的多账号管理工具。

**这份文档现在包含了所有核心仓库地址与技术路径，建议将其妥善保存在项目根目录的 `README.md` 或 `DOCS.md` 中。**

```

---


更新说明：

v2.9 优化界面逻辑

v2.4 优化流程逻辑

v2.0 原神登陆和切换功能