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

# --- 1. 确保管理员权限 (解决 0x0 坐标的核心) ---
def is_admin():
    try:
        return ctypes.windll.shell32.IsUserAnAdmin()
    except:
        return False

if not is_admin():
    print("?? 正在请求管理员权限以获取游戏窗口坐标...")
    ctypes.windll.shell32.ShellExecuteW(None, "runas", sys.executable, __file__, None, 1)
    sys.exit()

# --- 2. 配置路径 (指向你的 Go 项目 bin 目录) ---
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
OCR_EXE = os.path.join(BASE_DIR, "bin", "PaddleOCR-json_v1.4.1", "PaddleOCR-json.exe")
TMP_IMG = r"C:\Users\Public\sr_debug_temp.png" # 避开路径中文

def get_game_screenshot():
    # 支援简繁英窗口名
    targets = ["崩坏：星穹铁道", "崩坏：星穹铁道", "Star Rail"]
    hwnd = 0
    for t in targets:
        hwnd = win32gui.FindWindow(None, t)
        if hwnd: break

    if not hwnd:
        print("? 找不到星铁窗口，请确保游戏已运行！")
        return None

    # 强制还原窗口（解决最小化导致 0x0 的问题）
    if win32gui.IsIconic(hwnd):
        print("?? 检测到窗口最小化，正在尝试恢复...")
        win32gui.ShowWindow(hwnd, win32con.SW_RESTORE)
        time.sleep(1)

    # 带到前台
    try:
        win32gui.SetForegroundWindow(hwnd)
        time.sleep(1)
    except Exception as e:
        print(f"?? 无法置顶窗口: {e}")

    # 获取座标
    rect = win32gui.GetWindowRect(hwnd)
    x1, y1, x2, y2 = rect
    w, h = x2 - x1, y2 - y1
    
    if w <= 0 or h <= 0:
        print(f"? 错误：获取到的窗口尺寸为 {w}x{h}。请尝试手动点击一下游戏画面後再运行。")
        return None

    print(f"?? 成功锁定窗口: {w}x{h} 座标:({x1}, {y1})")

    # 截取全屏并裁切
    screenshot = pyautogui.screenshot()
    frame = cv2.cvtColor(np.array(screenshot), cv2.COLOR_RGB2BGR)
    
    img_h, img_w = frame.shape[:2]
    crop_y1, crop_y2 = max(0, y1), min(img_h, y2)
    crop_x1, crop_x2 = max(0, x1), min(img_w, x2)
    
    return frame[crop_y1:crop_y2, crop_x1:crop_x2]

def run_debug():
    frame = get_game_screenshot()
    if frame is None: return

    # 保存图片供 OCR 工具调用
    cv2.imwrite(TMP_IMG, frame)
    print(f"? 截图已保存至: {TMP_IMG}")

    if not os.path.exists(OCR_EXE):
        print(f"? 找不到 OCR 工具: {OCR_EXE}")
        return

    print(f"?? 正在调用 bin 目录下的 OCR 进行识别...")
    
    # 调用逻辑：必须切换 cwd 到 exe 目录，否则 DLL 会找不着
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
            print("? OCR 工具没有返回任何数据")
            if stderr: print(f"错误信息: {stderr}")
            return

        result = json.loads(stdout)
        print("\n" + "="*75)
        print(f"{'内容':<30} | {'中心X':<8} | {'中心Y':<8} | {'置信度':<8}")
        print("-" * 75)

        if result.get("code") == 100:
            for item in result.get("data", []):
                text = item.get("text")
                box = item.get("box")
                cx = (box[0][0] + box[2][0]) // 2
                cy = (box[0][1] + box[2][1]) // 2
                print(f"{text:<30} | {cx:<8} | {cy:<8} | {item.get('score'):.2f}")
        else:
            print(f"? OCR 识别返回错误: {result.get('msg')}")

    except Exception as e:
        print(f"? 执行出错: {e}")

    print("="*75)
    input("\n探测结束，按回车键退出...")

if __name__ == "__main__":
    # 自动安装基础库
    for lib in ["psutil", "pyautogui", "opencv-python", "numpy", "pywin32"]:
        try: __import__(lib)
        except ImportError: subprocess.check_call([sys.executable, "-m", "pip", "install", lib])
    
    run_debug()