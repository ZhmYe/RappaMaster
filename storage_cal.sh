#!/bin/bash

# 脚本名称：calculate_total_storage.sh
# 功能：统计 /hdd1 到 /hdd7 下所有 node_*/meta 目录的文件总大小
# 使用方法：./calculate_total_storage.sh

echo "开始统计各硬盘节点存储使用情况..."
echo "--------------------------------"

total_size=0  # 总大小（字节）
total_human=0 # 总大小（人类可读格式）

# 检查是否有权限访问 /hdd 目录
if [ ! -d "/hdd1" ]; then
  echo "错误：/hdd1 目录不存在，请确保脚本在有 /hdd1-7 目录的系统上运行"
  exit 1
fi

# 遍历 hdd1 到 hdd7
for hdd in {1..7}; do
  hdd_path="/hdd${hdd}"
  
  # 检查该硬盘目录是否存在
  if [ -d "$hdd_path" ]; then
    node_count=0
    hdd_size=0
    
    # 查找该硬盘下的所有 node_* 目录
    while IFS= read -r -d $'\0' node_path; do
      # 检查 meta 目录是否存在
      meta_path="${node_path}/meta"
      if [ -d "$meta_path" ]; then
        # 计算该 meta 目录大小
        node_size=$(du -sb "$meta_path" | awk '{print $1}')
        hdd_size=$((hdd_size + node_size))
        node_count=$((node_count + 1))
      fi
    done < <(find "$hdd_path" -maxdepth 1 -type d -name "node_*" -print0)
    
    # 转换为人类可读格式
    hdd_size_human=$(numfmt --to=iec $hdd_size)
    
    # 累加到总大小
    total_size=$((total_size + hdd_size))
    
    echo "硬盘 /hdd${hdd}:"
    echo "  节点数量: $node_count"
    echo "  总使用量: $hdd_size_human ($hdd_size 字节)"
    echo "--------------------------------"
  else
    echo "警告：/hdd${hdd} 目录不存在，跳过统计"
    echo "--------------------------------"
  fi
done

# 计算总大小的人类可读格式
total_human=$(numfmt --to=iec $total_size)

echo "统计完成！"
echo "========================================"
echo "所有硬盘节点总使用量: $total_human ($total_size 字节)"
echo "========================================"

exit 0