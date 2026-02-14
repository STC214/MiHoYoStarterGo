import os
from opencc import OpenCC


def convert_t2s(content):
    """将繁体字符串转换为简体"""
    # 't2s' 表示 Traditional Chinese to Simplified Chinese
    cc = OpenCC('t2s')
    return cc.convert(content)


def process_files(directory, extensions=('.md', '.go', '.vue', '.js', '.txt')):
    """遍历目录并转换指定后缀的文件"""
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith(extensions):
                file_path = os.path.join(root, file)

                # 读取内容
                with open(file_path, 'r', encoding='utf-8') as f:
                    try:
                        content = f.read()
                    except UnicodeDecodeError:
                        print(f"跳过无法解码的文件: {file_path}")
                        continue

                # 执行转换
                simplified_content = convert_t2s(content)

                # 写回文件
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(simplified_content)

                print(f"已完成转换: {file_path}")


if __name__ == "__main__":
    # 将此路径替换为你项目的实际路径
    project_path = './MiHoYoStarterGo'
    process_files(project_path)
