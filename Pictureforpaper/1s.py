# import seaborn as sns
# data = sns.load_dataset("tips")
# filepath = '/root/.pip'
# fig_name = 'scatterplot.png'

# # fig_path为想要存入的文件
# fig_path = filepath + '/' + fig_name
# fig = sns.scatterplot(x = data['total_bill'], y = data['tip'], hue = 'time', 
# data = data, palette = 'Set1', s = 100)
# scatter_fig = fig.get_figure()
# scatter_fig.savefig(fig_path, dpi = 400)

import seaborn as sns
import pandas as pd
import matplotlib.pyplot as plt
from matplotlib.ticker import FixedLocator, FixedFormatter
# import matplotlib.font_manager as font_manager
# Load an example datase
filepath = '/root/Reputation-based-state-machine-replication-/Pictureforpaper'
fig_name = 'blocksizesnew.svg'
fig_path = filepath + '/' + fig_name
# Create a visualization
# 加载数据
flights=pd.read_csv('ssa1.csv')
flights.head()
# 长型数据多折线图
dict1 = {'axes.axisbelow': True,  #轴在图形的下面
'axes.edgecolor': 'black',	#边框的颜色
'axes.facecolor': '#EAEAF2',	#背景颜色
'axes.grid': True,	#是否显示网格
'axes.labelcolor': '.15',
'axes.linewidth': 0.0,
'figure.facecolor': 'white',
'font.family': ['sans-serif'],
'font.sans-serif': ['Arial',
'Liberation Sans',
'Bitstream Vera Sans',
'sans-serif'],
'grid.color': 'white',
'grid.linestyle': '--',
'image.cmap': 'Greys',
'legend.frameon': True,
'legend.numpoints': 1,
'legend.scatterpoints': 1,
'lines.solid_capstyle': 'round',
'text.color': 'black',
'xtick.color': 'black',
'xtick.direction': 'out',
'xtick.major.size': 0.0,
'xtick.minor.size': 0.0,
'ytick.color': 'black',
'ytick.direction': 'out',
'ytick.major.size': 0.0,
'ytick.minor.size': 0.0
}
sns.set(context='paper', 
palette='deep', font='sans-serif', font_scale=1, 
color_codes=True, rc=dict1)
sns.despine(fig=None, ax=None,
 top=True, right=True, left=True, bottom=True,
  offset=None, trim=False)
palette = sns.color_palette("bright")
fig = sns.lineplot(data=flights,x='吞吐量/tps',y='共识时延/ms',dashes=False,sort=True,
errorbar=None,hue='区块大小',style='区块大小',markers=['^','D','v'] ,linewidth =1.5,
orient='y', markeredgecolor = 'none',alpha = 1)
#alpha 设置透明度
#mark空心，mark见matlop
#markerfacecolor='none'
# fig.set_xlim(left=0, right=900)  # 确保x轴从0开始
# fig.set_ylim(bottom=1000, top=1500)  # 确保y轴从980开始
fig.set_xlim(left=0, right=900)  # 确保x轴从0开始
fig.set_ylim(bottom=960, top=1500)  # 确保y轴从980开始
# plt.scatter([5], [700], color='black')  # 在(0, 980)处添加一个红色的点标记
# plt.text(5, 700, '5, 700', color='black', ha='right')  # 添加文本说明
# 设定特定的x轴刻度位置
plt.gca().xaxis.set_major_locator(FixedLocator([0] + list(plt.gca().get_xticks())))
# 设定特定的y轴刻度位置
plt.gca().yaxis.set_major_locator(FixedLocator([960] + list(plt.gca().get_yticks())))
#设置x,y轴label大小
# fig.xaxis.label.set_size(15)
#科学计数法
plt.rcParams['font.sans-serif'] = ['SimHei']

fig.xaxis.label.set_size(14)
fig.yaxis.label.set_size(14)
#设zhi刻度大小
plt.xticks(fontsize=15, rotation=0)
plt.yticks(fontsize=15, rotation=0)
#设置标签大小
# fig.set_axis_labels(fontsize=20)
#科学计数法
# plt.yscale('log')
plt.setp(fig.get_legend().get_texts(), fontsize='12') # for legend text
plt.setp(fig.get_legend().get_title(), fontsize='12') # for legend title
lineplot_figure = fig.get_figure()
lineplot_figure.savefig(fig_path, dpi = 1000)
plt.show()
