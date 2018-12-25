#! /usr/bin/env python
# -*- coding: utf-8 -*-
#切换到老主干trunk：python ChangeSvn.py trunk 

import sys,os

svn_branch_date = sys.argv[1]

# 定义主干svn业务层和框架层路径
svn_trunk = 'svn://192.168.150.63/timefire/trunk/server'
new_svn_trunk = 'svn://192.168.150.63/timefire/newtrunk/server'
zeus_trunk = 'svn://192.168.150.63/mb/zeus/branch_1'
new_zeus_trunk = 'svn://192.168.150.63/mb/zeus/newtrunk'

# 定义分支svn业务层和框架层路径
svn_branch = 'svn://192.168.150.63/timefire/tags/branche_' + svn_branch_date + '_pb/server'
zeus_branch = 'svn://192.168.150.63/mb/zeus/tags/branche_' + svn_branch_date + '_pb'

# 删除旧的工程文件,首先要改变工作目录

print('准备开始切换服务器工程！！稍等片刻')

svn_path = os.getcwd() + os.path.sep +'svn'
os.chdir(svn_path)
os.system('rm -rf server/')
os.system('rm -rf branch_1')

if svn_branch_date == 'trunk':
    os.system('svn co {0}'.format(svn_trunk))
    os.system('svn co {0} branch_1'.format(zeus_trunk))
    print("切换完毕，启动服务器之前记得修改publish.sh文件里面的ip地址哦！！")
elif svn_branch_date == 'newtrunk':
    os.system('svn co {0}'.format(new_svn_trunk))
    os.system('svn co {0} newtrunk'.format(new_zeus_trunk))
    print("切换完毕，启动服务器之前记得修改publish.sh文件里面的ip地址哦！！")
elif svn_branch_date.isdigit() and (len(svn_branch_date) == 8):
    os.system('svn co {0}'.format(svn_branch))
    os.system('svn co {0} branch_1'.format(zeus_branch))
    print("切换完毕，启动服务器之前记得修改publish.sh文件里面的ip地址哦！！")
else:
    print('哎呀！切换失败了！请查看下面的说明哦！')
    print('该脚本目前仅支持切换主干以及格式常规的分支')
    print('''输入的格式需要符合以下条件：
    ********************************************
    1、切到主干的话参数是trunk
    2、切到分支的话是对应的分支日期，格式如20180106
    ********************************************''')
