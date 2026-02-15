import os
import sys
import json
import time
import subprocess
import ctypes
import cv2
import numpy as np
import pyautogui
import win32gui
import win32con

# --- 1. 確保管理員權限 (解決 0x0 坐標的核心) ---
def is_admin():
    try:
        return ctypes.windll.shell32.IsUserAnAdmin()
    except:
        return False

if not is_admin():
    print("🚀 正在請求管理員權限以獲取遊戲窗口坐標...")
    ctypes.windll.shell32.ShellExecuteW(None, "runas", sys.executable, __file__, None, 1)
    sys.exit()

# --- 2. 配置路徑 (指向你的 Go 項目 bin 目錄) ---
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
OCR_EXE = os.path.join(BASE_DIR, "bin", "PaddleOCR-json_v1.4.1", "PaddleOCR-json.exe")
TMP_IMG = r"C:\Users\Public\sr_debug_temp.png" # 避開路徑中文

def get_game_screenshot():
    # 支援簡繁英窗口名
    targets = ["崩坏：星穹铁道", "崩壞：星穹鐵道", "Star Rail"]
    hwnd = 0
    for t in targets:
        hwnd = win32gui.FindWindow(None, t)
        if hwnd: break

    if not hwnd:
        print("❌ 找不到星鐵窗口，請確保遊戲已運行！")
        return None

    # 強制還原窗口（解決最小化導致 0x0 的問題）
    if win32gui.IsIconic(hwnd):
        print("📢 檢測到窗口最小化，正在嘗試恢復...")
        win32gui.ShowWindow(hwnd, win32con.SW_RESTORE)
        time.sleep(1)

    # 帶到前台
    try:
        win32gui.SetForegroundWindow(hwnd)
        time.sleep(1)
    except Exception as e:
        print(f"⚠️ 無法置頂窗口: {e}")

    # 獲取座標
    rect = win32gui.GetWindowRect(hwnd)
    x1, y1, x2, y2 = rect
    w, h = x2 - x1, y2 - y1
    
    if w <= 0 or h <= 0:
        print(f"❌ 錯誤：獲取到的窗口尺寸為 {w}x{h}。請嘗試手動點擊一下遊戲畫面後再運行。")
        return None

    print(f"🎯 成功鎖定窗口: {w}x{h} 座標:({x1}, {y1})")

    # 截取全屏並裁切
    screenshot = pyautogui.screenshot()
    frame = cv2.cvtColor(np.array(screenshot), cv2.COLOR_RGB2BGR)
    
    img_h, img_w = frame.shape[:2]
    crop_y1, crop_y2 = max(0, y1), min(img_h, y2)
    crop_x1, crop_x2 = max(0, x1), min(img_w, x2)
    
    return frame[crop_y1:crop_y2, crop_x1:crop_x2]

def run_debug():
    frame = get_game_screenshot()
    if frame is None: return

    # 保存圖片供 OCR 工具調用
    cv2.imwrite(TMP_IMG, frame)
    print(f"✅ 截圖已保存至: {TMP_IMG}")

    if not os.path.exists(OCR_EXE):
        print(f"❌ 找不到 OCR 工具: {OCR_EXE}")
        return

    print(f"🔍 正在調用 bin 目錄下的 OCR 進行識別...")
    
    # 調用邏輯：必須切換 cwd 到 exe 目錄，否則 DLL 會找不著
    args = [OCR_EXE, f"--image_path={TMP_IMG}"]
    try:
        proc = subprocess.Popen(
            args, 
            cwd=os.path.dirname(OCR_EXE),
            stdout=subprocess.PIPE, 
            stderr=subprocess.PIPE,
            text=True,
            encoding='utf-8',
            errors='ignore'
        )
        stdout, stderr = proc.communicate()
        
        if not stdout:
            print("❌ OCR 工具沒有返回任何數據")
            if stderr: print(f"錯誤信息: {stderr}")
            return

        result = json.loads(stdout)
        print("\n" + "="*75)
        print(f"{'內容':<30} | {'中心X':<8} | {'中心Y':<8} | {'置信度':<8}")
        print("-" * 75)

        if result.get("code") == 100:
            for item in result.get("data", []):
                text = item.get("text")
                box = item.get("box")
                cx = (box[0][0] + box[2][0]) // 2
                cy = (box[0][1] + box[2][1]) // 2
                print(f"{text:<30} | {cx:<8} | {cy:<8} | {item.get('score'):.2f}")
        else:
            print(f"❌ OCR 識別返回錯誤: {result.get('msg')}")

    except Exception as e:
        print(f"❌ 執行出錯: {e}")

    print("="*75)
    input("\n探測結束，按回車鍵退出...")

if __name__ == "__main__":
    # 自動安裝基礎庫
    for lib in ["psutil", "pyautogui", "opencv-python", "numpy", "pywin32"]:
        try: __import__(lib)
        except ImportError: subprocess.check_call([sys.executable, "-m", "pip", "install", lib])
    
    run_debug()