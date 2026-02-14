---
slug: posts/2026/README
title: "項目技術總結報告：MiHoYoStarterGo"
description: "就是个原神多账号切换器"
date: 2026-02-14
author: "STC"
tags: ["NOTE"]
featured: false
editable: true
---

---

# <span style="color:red">免责声明：工具所有数据仅你本机存储，不含任何网络数据传输功能（毕竟连自动检测更新都没做），若遇到需要备份的情况请将工具目录整体打包后放在你认为丢不了的地方。恢复的时候只需要解压到非中文和特殊符号目录下即可。
其次，工具不含有任何外挂行为，不含有任何内存操作。</span>


# 項目技術總結報告：MiHoYoStarterGo (開發者存檔版)

## 项目可执行文件刚需PaddleOCR-Go。如需使用请于下方介绍中下载对应版本。目前项目使用1.4.1版。
## 功能目前只做了原神部分，其他部分待完善。

## 1. 核心 OCR 技術選型

本項目的圖像識別能力建立在開源社區優秀的離線推理庫之上。

* **技術倉庫**：[otiai10/gosseract](https://github.com/otiai10/gosseract) (基於 Tesseract) 或 **[PaddleOCR-Go](https://github.com/otiai10/gosseract)** (推薦參考)。
* *註：根據我們目前的實現，主要依賴於本地編譯的 PaddleOCR 推理引擎，確保了對中文（簡/繁）極高的識別準確率。*


* **倉庫簡介**：PaddleOCR 是百度開源的超輕量級 OCR 系統，支持 80 多種語言，其特點是模型小（僅幾 M）、速度快、精度高，非常適合嵌入到像 Wails 這樣的桌面應用中進行離線推斷。

---

## 2. 離線 OCR 的技術優勢

* **隱私與安全**：所有截圖處理均在本地 `temp_*.png` 文件中完成，並在協程結束時由 `os.Remove` 徹底清理。
* **零成本運行**：無需申請雲端 API Key，沒有調用次數限制，且在無網絡環境下（如飛機、高鐵）依然能正常工作。
* **環境干擾排除**：開發過程中發現，若使用小米澎湃系統，需關閉**「暈車緩解功能」**，否則系統級的視覺補償會干擾截圖與坐標定位。

---

## 3. 已實現的核心功能閉環

| 模塊 | 技術重點 | 關鍵函數 / 源代碼 |
| --- | --- | --- |
| **自動化監控** | 異步循環檢測、窗口置頂、截圖識別 | `StartAutomationMonitor` |
| **交互控制** | 支持全局暫停、手動停止監控並回收資源 | `TogglePause`, `StopMonitor` |
| **賬號管理** | 密碼加密存儲、Token 提取、賬號刪除與同步 | `AddAccount`, `DeleteAccount` |
| **前端 UI** | Vue3 響應式佈局、多主題切換、實時監控彈窗 | `App.vue` |

---

## 4. 項目部署與權限聲明

為了保證離線 OCR 和註冊表讀取的權限，打包時必須包含以下設置：

1. **管理員清單**：`build/windows/wails.exe.manifest` 中已配置 `requireAdministrator` 以獲取系統句柄。
2. **打包命令**：
```bash
wails build -clean -platform windows/amd64 -ldflags "-s -w"

```


3. **依賴說明**：如果更換開發環境，請確保 `gosseract` 依賴的 `tesseract-ocr` 庫或 `Paddle` 推理庫已正確配置於系統環境變量中。

---

## 5. 項目總結

從最初的 Go 後端邏輯設計，到 Vue3 前端 UI 的打磨，再到**離線 PaddleOCR** 的深度集成與 **Windows UAC 權限** 的攻克，**MiHoYoStarterGo** 現在已成為一個安全、高效、且具備工業級穩定性的多賬號管理工具。

**這份文檔現在包含了所有核心倉庫地址與技術路徑，建議將其妥善保存在項目根目錄的 `README.md` 或 `DOCS.md` 中。**
