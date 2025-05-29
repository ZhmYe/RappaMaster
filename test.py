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
    plt.savefig("node_scalability.svg", bbox_inches='tight')
    plt.close()