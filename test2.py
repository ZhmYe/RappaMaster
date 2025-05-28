import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from matplotlib.ticker import MaxNLocator

# ================== 全局样式设置 ==================
plt.style.use('seaborn-paper')
sns.set_palette("tab10")
plt.rcParams.update({
    "font.family": "Times New Roman",
    "font.weight": "bold",
    "font.size": 11,
    "axes.labelweight": "bold",
    "axes.titlesize": 0,  # 禁用标题
    "axes.labelsize": 10,
    "xtick.labelsize": 9,
    "ytick.labelsize": 9,
    "figure.dpi": 300,
    "figure.figsize": (5, 3.5),
    "grid.linestyle": "--",
    "grid.alpha": 0.3,
    "svg.fonttype": "none"
})

# ================== 图表1: 数据收集效率 ==================
def plot_collection():
    data = {
        'Size': [1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 11000, 12000, 13000, 14000, 15000],
        'collect': [0.1, 0.19, 0.31, 0.39, 0.54, 0.69, 0.74, 0.76, 0.94, 1.02, 1.2, 1.21, 1.23, 1.42, 1.51]
    }
    df = pd.DataFrame(data)

    fig, ax = plt.subplots()
    sns.lineplot(x='Size', y='collect', data=df, marker='o', 
                linewidth=1.2, markersize=5, color='#1f77b4')
    
    ax.set(xlabel='Data Size (rows)', ylabel='Collection Time (s)')
    ax.xaxis.set_major_locator(MaxNLocator(6))
    ax.yaxis.set_major_locator(MaxNLocator(6))
    
    plt.tight_layout()
    plt.savefig("data_collection.pdf", bbox_inches='tight')
    plt.close()

# ================== 图表2: 处理吞吐量 ================== 
def plot_processing():
    data = {
        'Size': [1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 11000, 12000, 13000, 14000, 15000],
        'duration': [22.693, 23.197, 23.614, 23.484, 23.177, 63.102, 63.09, 63.416, 62.961, 62.412, 
                    107.414, 106.908, 106.832, 106.579, 105.968]
    }
    df = pd.DataFrame(data)

    fig, ax = plt.subplots()
    sns.lineplot(x='Size', y='duration', data=df, marker='s',
                linewidth=1.2, markersize=5, color='#d62728')
    
    ax.set(xlabel='Data Size (rows)', ylabel='Processing Time (s)')
    ax.xaxis.set_major_locator(MaxNLocator(6))
    ax.yaxis.set_major_locator(MaxNLocator(6))
    
    # 添加阶段标记
    ax.axvline(6000, color='grey', linestyle='--', alpha=0.6)
    ax.axvline(11000, color='grey', linestyle='--', alpha=0.6)
    ax.text(3500, 100, "Schedule 1", ha='center', fontstyle='italic')
    ax.text(8500, 100, "Schedule 2", ha='center', fontstyle='italic')
    ax.text(12500, 100, "Schedule 3", ha='center', fontstyle='italic')
    
    plt.tight_layout()
    plt.savefig("data_processing.pdf", bbox_inches='tight')
    plt.close()

# ================== 图表3: 节点扩展性 ==================
def plot_scalability():
    data = {
        "Node Count": [1, 5, 10, 15, 20, 25, 30],
        "Execution Time (s)": [186.330, 25.127, 24.986, 63.158, 70.540, 77.215, 93.975]
    }
    df = pd.DataFrame(data)

    fig, ax = plt.subplots()
    sns.lineplot(x="Node Count", y="Execution Time (s)", data=df,
                marker='D', color='#2ca02c', linewidth=1.2, markersize=6)
    
    # 修改坐标轴设置
    ax.set(xlabel="Number of Nodes", ylabel="Processing Time (s)")  # 修改纵坐标标签
    ax.set_xticks([1, 5, 10, 15, 20, 25, 30])  # 明确设置所有节点刻度
    ax.set_xticklabels([1, 5, 10, 15, 20, 25, 30])  # 确保显示所有刻度标签
    
    # 最佳性能区间
    ax.axvspan(4, 11, color='#ffd700', alpha=0.15)
    ax.text(7.5, 170, "Optimal Range", ha='center', fontstyle='italic')
    
    plt.tight_layout()
    plt.savefig("node_scalability.pdf", bbox_inches='tight')
    plt.close()

# ================== 执行生成 ==================
if __name__ == "__main__":
    plot_collection()
    plot_processing()
    plot_scalability()
    print("三张学术图表生成完成，已保存为SVG格式")