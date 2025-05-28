import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from matplotlib.ticker import FuncFormatter, LogLocator
import numpy as np

# ================== 全局样式设置 ==================
plt.style.use('seaborn-paper')
sns.set_palette("tab10")
plt.rcParams.update({
    "font.family": "Times New Roman",
    "font.weight": "bold",
    "axes.labelweight": "bold",
    "axes.titlesize": 0,
    "grid.linestyle": "--",
    "grid.alpha": 0.3,
    "figure.dpi": 300,
    "figure.figsize": (5, 3.5),
    "svg.fonttype": "none"
})

# ================== 数据准备 ==================
data = [
    [100000, 20817, 4803.77, 4896, 20424.84],
    [150000, 35778, 4192.52, 7325, 20477.82],
    [200000, 52732, 3792.76, 8902, 22466.86],
    [250000, 62578, 3995.01, 11234, 22253.87],
    [300000, 73674, 4071.99, 13726, 21856.33],
    [350000, 69054, 5068.50, 18337, 19087.09],
    [400000, 103869, 3851.00, 23639, 16921.19]
]

df = pd.DataFrame(data, columns=[
    'Evidence Count', 'Time Single', 'Speed Single',
    'Time Batch', 'Speed Batch'
])

# ================== 存证速度分析图表 ==================
def plot_speed():
    melt_df = df.melt(id_vars='Evidence Count', 
                     value_vars=['Speed Batch', 'Speed Single'],
                     var_name='Mode', value_name='Speed')
    melt_df['Mode'] = melt_df['Mode'].replace({
        'Speed Batch': 'Batch Processing',
        'Speed Single': 'Single Processing'
    })

    fig, ax = plt.subplots()
    sns.lineplot(x='Evidence Count', y='Speed', hue='Mode', style='Mode',
                data=melt_df, markers={'Batch Processing':'o', 'Single Processing':'s'},
                linewidth=1.5, markersize=8, dashes=False)
    
    # 坐标轴优化
    ax.xaxis.set_major_formatter(FuncFormatter(lambda x, _: f'{int(x/1000)}k'))
    ax.yaxis.set_major_formatter(FuncFormatter(lambda x, _: f'{x/1000:.1f}k' if x >= 1000 else f'{int(x)}'))
    
    # 设置y轴标签（添加单位）
    ax.set_ylabel('Throughput (evidence/s)')
    
    # 调整图例位置和边距
    plt.legend(title='', loc='upper center', bbox_to_anchor=(0.5, 1.15), 
              ncol=2, frameon=True)
    plt.tight_layout(rect=[0, 0, 1, 0.95])
    plt.savefig("updated_speed_analysis.pdf")
    plt.close()

# ================== 处理耗时分析图表 ==================
def plot_latency():
    melt_df = df.melt(id_vars='Evidence Count', 
                     value_vars=['Time Batch', 'Time Single'],
                     var_name='Mode', value_name='Latency')
    melt_df['Mode'] = melt_df['Mode'].replace({
        'Time Batch': 'Batch Processing',
        'Time Single': 'Single Processing'
    })

    fig, ax = plt.subplots()
    sns.lineplot(x='Evidence Count', y='Latency', hue='Mode', style='Mode',
                data=melt_df, markers={'Batch Processing':'o', 'Single Processing':'s'},
                linewidth=1.5, markersize=8, dashes=False)
    
    # 坐标轴优化（线性刻度）
    ax.xaxis.set_major_formatter(FuncFormatter(lambda x, _: f'{int(x/1000)}k'))
    
    # 设置 y 轴刻度为线性（均匀分布）
    max_latency = melt_df['Latency'].max()
    y_ticks = np.linspace(0, max_latency * 1.1, num=6)  # 6 个均匀分布的刻度
    ax.set_yticks(y_ticks)
    
    # 自定义 y 轴标签格式（k 表示千）
    def linear_fmt(x, pos):
        if x >= 1000:
            return f'{x/1000:.0f}k'
        else:
            return f'{int(x)}'
    
    ax.yaxis.set_major_formatter(FuncFormatter(linear_fmt))
    
    # 设置 y 轴标签（添加单位）
    ax.set_ylabel('Latency (ms)')
    
    # 调整图例位置和边距
    plt.legend(title='', loc='upper center', bbox_to_anchor=(0.5, 1.15),
              ncol=2, frameon=True)
    plt.tight_layout(rect=[0, 0, 1, 0.95])
    plt.savefig("updated_latency_analysis_linear.pdf")
    plt.close()

# ================== 执行生成 ==================
if __name__ == "__main__":
    plot_speed()
    plot_latency()