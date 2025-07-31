import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Statistic, 
  Button, 
  Space, 
  List, 
  Tag,
  Progress,
  Typography
} from 'antd';
import { 
  RobotOutlined, 
  FileTextOutlined, 
  ToolOutlined,
  PlayCircleOutlined,
  PlusOutlined,
  MessageOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ApiService, { Agent, Session, Task } from '../services/api';

const { Title, Paragraph } = Typography;

const Home: React.FC = () => {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    setLoading(true);
    try {
      const [agentsData, sessionsData, tasksData] = await Promise.all([
        ApiService.listAgents(),
        ApiService.listSessions(),
        ApiService.listTasks()
      ]);
      setAgents(agentsData);
      setSessions(sessionsData);
      setTasks(tasksData);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const getRecentTasks = () => {
    return tasks
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 5);
  };

  const getActiveSessions = () => {
    return sessions.filter(session => session.status === 'active');
  };

  const getRunningAgents = () => {
    return agents.filter(agent => agent.state === 'running');
  };

  const getTaskStatusStats = () => {
    const stats = {
      pending: 0,
      running: 0,
      completed: 0,
      failed: 0
    };
    
    tasks.forEach(task => {
      if (stats.hasOwnProperty(task.status)) {
        stats[task.status as keyof typeof stats]++;
      }
    });
    
    return stats;
  };

  const taskStats = getTaskStatusStats();

  return (
    <div>
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>欢迎使用 Coco AI Research</Title>
        <Paragraph>
          智能研究平台，让AI助手帮助您进行深度研究和分析
        </Paragraph>
      </div>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="智能体总数"
              value={agents.length}
              prefix={<RobotOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃会话"
              value={getActiveSessions().length}
              prefix={<MessageOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="运行中任务"
              value={taskStats.running}
              prefix={<PlayCircleOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="可用工具"
              value={5}
              prefix={<ToolOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 快速操作 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={24}>
          <Card title="快速操作">
            <Space size="large">
              <Button 
                type="primary" 
                icon={<PlusOutlined />}
                onClick={() => navigate('/agents')}
              >
                创建智能体
              </Button>
              <Button 
                icon={<MessageOutlined />}
                onClick={() => navigate('/sessions')}
              >
                开始研究会话
              </Button>
              <Button 
                icon={<PlayCircleOutlined />}
                onClick={() => navigate('/tasks')}
              >
                查看任务
              </Button>
              <Button 
                icon={<ToolOutlined />}
                onClick={() => navigate('/tools')}
              >
                使用工具
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>

      {/* 任务状态 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={12}>
          <Card title="任务状态统计">
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <span>等待中: {taskStats.pending}</span>
                <Progress percent={tasks.length > 0 ? (taskStats.pending / tasks.length) * 100 : 0} size="small" />
              </div>
              <div>
                <span>执行中: {taskStats.running}</span>
                <Progress percent={tasks.length > 0 ? (taskStats.running / tasks.length) * 100 : 0} size="small" status="active" />
              </div>
              <div>
                <span>已完成: {taskStats.completed}</span>
                <Progress percent={tasks.length > 0 ? (taskStats.completed / tasks.length) * 100 : 0} size="small" status="success" />
              </div>
              <div>
                <span>失败: {taskStats.failed}</span>
                <Progress percent={tasks.length > 0 ? (taskStats.failed / tasks.length) * 100 : 0} size="small" status="exception" />
              </div>
            </Space>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="运行中智能体">
            <List
              size="small"
              dataSource={getRunningAgents()}
              renderItem={(agent) => (
                <List.Item>
                  <Space>
                    <RobotOutlined style={{ color: '#52c41a' }} />
                    <span>{agent.name}</span>
                    <Tag color="green">{agent.state}</Tag>
                  </Space>
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      {/* 最近任务 */}
      <Row gutter={16}>
        <Col span={24}>
          <Card title="最近任务" loading={loading}>
            <List
              size="small"
              dataSource={getRecentTasks()}
              renderItem={(task) => (
                <List.Item
                  actions={[
                    <Button 
                      type="link" 
                      size="small"
                      onClick={() => navigate(`/tasks/${task.id}`)}
                    >
                      查看详情
                    </Button>
                  ]}
                >
                  <List.Item.Meta
                    title={
                      <Space>
                        <span>{task.query}</span>
                        <Tag color={
                          task.status === 'completed' ? 'green' :
                          task.status === 'running' ? 'blue' :
                          task.status === 'failed' ? 'red' : 'default'
                        }>
                          {task.status}
                        </Tag>
                      </Space>
                    }
                    description={`创建时间: ${new Date(task.created_at).toLocaleString()}`}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Home; 