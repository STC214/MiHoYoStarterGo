import os

def generate_tree(startpath, output_file):
    # 需要忽略的文件夹，避免干扰（比如 .git, __pycache__, node_modules 等）
    ignore_dirs = {'.git', '__pycache__', '.idea', '.vscode', 'vendor', 'obj', 'bin_temp'}
    
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(f"?? Project Structure for: {os.path.abspath(startpath)}\n")
        f.write("=" * 50 + "\n")
        
        for root, dirs, files in os.walk(startpath):
            # 过滤掉不需要的目录
            dirs[:] = [d for d in dirs if d not in ignore_dirs]
            
            level = root.replace(startpath, '').count(os.sep)
            indent = '│   ' * (level)
            f.write(f'{indent}├── {os.path.basename(root)}/\n')
            
            sub_indent = '│   ' * (level + 1)
            for file in files:
                if file != output_file and file != 'gen_tree.py': # 不记录生成的报告和脚本本身
                    f.write(f'{sub_indent}└── {file}\n')

if __name__ == "__main__":
    output = "project_structure.txt"
    generate_tree('.', output)
    print(f"? 目录结构已生成到: {output}")