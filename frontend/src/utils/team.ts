// 关于我们页的团队成员数据
export interface TeamMember {
  avatar: string
  role: string
  username: string
  grade: string
  name: string
  bio: string
}

export const teamMembers: TeamMember[] = [
  {
    avatar: '/assets/avatar/XSG.png',
    role: '「网站开发者」',
    username: '小帅哥',
    grade: '23软件工程',
    name: '林恬烁',
    bio: '学校首支ACM区域赛金牌获奖队伍选手、学校首项大学生团体程序设计天梯赛个人国一获得者、前竞赛中心负责人、Acking 网站开发者。目前就职同花顺 摘星计划。',
  },
  {
    avatar: '/assets/avatar/yjddb.jpg',
    role: '「实验室创始人」',
    username: 'YJDDB',
    grade: '22计科',
    name: '代金宇',
    bio: 'AcKing实验室创始人，AchoBeta实验室话事人、前竞赛中心负责人。曾负责安排管理校内各算法比赛，为校内算法竞赛氛围做出很大贡献。目前就职字节跳动。',
  },
  {
    avatar: '/assets/avatar/ccx.jpg',
    role: '「指导老师」',
    username: '陈老',
    grade: '',
    name: '陈传祥 老师',
    bio: 'ACE实验室、AcKing实验室、AchoBeta实验室指导老师。指导管理校内的ACM竞赛等算法竞赛事项，并给予了我们很大支持和帮助。',
  },
  {
    avatar: '/assets/avatar/hh.jpg',
    role: '「传奇选手」',
    username: 'SeaYellow',
    grade: '22大数据',
    name: '黄海',
    bio: '学校首支ACM区域赛金牌获奖队伍队长、Codeforces 黄名、CCF-CSP认证435分全国前0.34%。目前就读德国卡尔斯鲁厄理工学院。',
  },
  {
    avatar: '/assets/avatar/fang_baby.png',
    role: '「传奇选手」',
    username: 'fang_baby',
    grade: '23大数据',
    name: '林志龙',
    bio: '学校首支ACM区域赛金牌获奖队伍选手、Codeforces 黄名。目前往信奥赛教培方向发展。',
  },
  {
    avatar: '/assets/avatar/ESTZ.jpg',
    role: '「传奇选手」',
    username: 'ESTZ',
    grade: '26计算机类',
    name: '张宇铭',
    bio: 'AcKing实验室当前负责人、大一期间达到 Codeforces 紫名、夺得邀请赛银牌、蓝桥杯国一。',
  },
]

// 使命板块
export const missions = [
  {
    title: '算法为何',
    icon: 'fa-lightbulb',
    circle: 'bg-blue-500',
    card: 'bg-blue-50',
    desc: '解决一个问题，往往有多种方法，但好的算法可以让问题的解决更加高效、准确。算法竞赛便是通过对题目的分析思考，选出最优的算法，在规定时间内解决问题。',
  },
  {
    title: '备赛意义',
    icon: 'fa-key',
    circle: 'bg-yellow-500',
    card: 'bg-yellow-50',
    desc: '算法竞赛是一项公平且有趣的竞赛，想取得成绩需要靠自身的天赋与努力。取得优异的奖项也会对就业、考研等方面有着巨大的帮助。',
  },
  {
    title: '团队目标',
    icon: 'fa-users',
    circle: 'bg-red-500',
    card: 'bg-red-50',
    desc: 'AcKing 算法竞赛实验室，致力于为校内算法竞赛爱好者提供一个学习、交流、竞赛的平台。为学校 ACM 竞赛圈的发展贡献自己的力量。',
  },
]
