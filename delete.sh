#!/bin/bash

# 脚本名称：delete_nodes.sh
# 功能：安全删除 /hdd1 到 /hdd7 下所有 node_* 目录
# 使用方法：./delete_nodes.sh [确认密码]

# 安全验证密码（防止误操作）
SAFETY_PASSWORD="confirm123"

if [ "$1" != "$SAFETY_PASSWORD" ]; then
  echo "================================================"
  echo "安全警告：此脚本将删除所有硬盘上的节点数据！"
  echo "================================================"
  echo "如需继续，请执行: $0 $SAFETY_PASSWORD"
  echo "================================================"
  exit 1
fi

echo "开始删除各硬盘节点数据..."
echo "--------------------------------"

total_deleted=0  # 删除的节点总数

# 遍历 hdd1 到 hdd7
for hdd in {1..7}; do
  hdd_path="/hdd${hdd}"
  
  # 检查该硬盘目录是否存在
  if [ -d "$hdd_path" ]; then
    node_count=0
    
    # 查找并删除该硬盘下的所有 node_* 目录
    while IFS= read -r -d $'\0' node_path; do
      echo "正在删除: $node_path"
      rm -rf "$node_path"
      if [ ! -d "$node_path" ]; then
        node_count=$((node_count + 1))
      else
        echo "警告: 未能成功删除 $node_path"
      fi
    done < <(find "$hdd_path" -maxdepth 1 -type d -name "node_*" -print0)
    
    total_deleted=$((total_deleted + node_count))
    echo "硬盘 /hdd${hdd}: 已删除 $node_count 个节点"
    echo "--------------------------------"
  else
    echo "通知：/hdd${hdd} 目录不存在，跳过处理"
    echo "--------------------------------"
  fi
done

echo "删除操作完成！"
echo "========================================"
echo "总共删除的节点数量: $total_deleted"
echo "========================================"

# 最后确认所有节点是否已删除
echo "最终检查节点目录是否存在..."
remaining_nodes=$(find /hdd{1..7} -maxdepth 1 -type d -name "node_*" 2>/dev/null | wc -l)
if [ "$remaining_nodes" -eq 0 ]; then
  echo "验证通过：所有节点目录已成功删除"
else
  echo "警告：仍有 $remaining_nodes 个节点目录未被删除"
  find /hdd{1..7} -maxdepth 1 -type d -name "node_*" 2>/dev/null
fi

exit 0