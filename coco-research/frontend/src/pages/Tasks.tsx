import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Button, 
  Space, 
  Tag, 
  Progress,
  message,
  Tooltip,
  Badge
} from 'antd';
import { 
  EyeOutlined, 
  DeleteOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  StopOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ApiService, { Task, Agent } from '../services/api';

const Tasks: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    loadTasks();
    loadAgents();
  }, []);

  const loadTasks = async () => {
    setLoading(true);
    try {
      const data = await ApiService.listTasks();
      setTasks(data);
    } catch (error) {
      message.error('加载任务列表失败');
      console.error('Failed to load tasks:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadAgents = async () => {
    try {
      const data = await ApiService.listAgents();
      setAgents(data);
    } catch (error) {
      console.error('Failed to load agents:', error);
    }
  };

  const handleViewTask = (taskId: string) => {
    navigate(`/tasks/${taskId}`);
  };

  const handleDeleteTask = async (taskId: string) => {
    try {
      await ApiService.deleteTask(taskId);
      message.success('任务删除成功');
      loadTasks();
    } catch (error) {
      message.error('删除任务失败');
      console.error('Failed to delete task:', error);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'processing';
      case 'completed':
        return 'success';
      case 'failed':
        return 'error';
      case 'cancelled':
        return 'default';
      default:
        return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending':
        return '等待中';
      case 'running':
        return '执行中';
      case 'completed':
        return '已完成';
      case 'failed':
        return '失败';
      case 'cancelled':
        return '已取消';
      default:
        return status;
    }
  };

  const getProgress = (status: string) => {
    switch (status) {
      case 'pending':
        return 0;
      case 'running':
        return 50;
      case 'completed':
        return 100;
      case 'failed':
        return 0;
      case 'cancelled':
        return 0;
      default:
        return 0;
    }
  };

  const columns = [
    {
      title: '任务ID',
      dataIndex: 'id',
      key: 'id',
      render: (id: string) => (
        <code style={{ fontSize: '12px' }}>{id.slice(0, 8)}...</code>
      ),
    },
    {
      title: '查询内容',
      dataIndex: 'query',
      key: 'query',
      ellipsis: true,
      render: (query: string) => (
        <Tooltip title={query}>
          <span>{query}</span>
        </Tooltip>
      ),
    },
    {
      title: '关联智能体',
      dataIndex: 'agent_id',
      key: 'agent_id',
      render: (agentId: string) => {
        const agent = agents.find(a => a.id === agentId);
        return agent ? (
          <Tag color="blue">{agent.name}</Tag>
        ) : (
          <Tag color="default">未知</Tag>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Space>
          <Badge status={getStatusColor(status)} />
          <Tag color={getStatusColor(status) === 'processing' ? 'blue' : 
                      getStatusColor(status) === 'success' ? 'green' : 
                      getStatusColor(status) === 'error' ? 'red' : 'default'}>
            {getStatusText(status)}
          </Tag>
        </Space>
      ),
    },
    {
      title: '进度',
      key: 'progress',
      render: (_: any, record: Task) => (
        <Progress 
          percent={getProgress(record.status)} 
          size="small"
          status={record.status === 'failed' ? 'exception' : undefined}
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Task) => (
        <Space>
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewTask(record.id)}
            />
          </Tooltip>
          
          {record.status === 'running' && (
            <Tooltip title="停止任务">
              <Button
                type="text"
                danger
                icon={<StopOutlined />}
                onClick={() => {
                  message.info('停止任务功能待实现');
                }}
              />
            </Tooltip>
          )}
          
          {record.status === 'pending' && (
            <Tooltip title="重新执行">
              <Button
                type="text"
                icon={<PlayCircleOutlined />}
                onClick={() => {
                  message.info('重新执行功能待实现');
                }}
              />
            </Tooltip>
          )}
          
          {record.status !== 'running' && (
            <Tooltip title="删除任务">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                onClick={() => handleDeleteTask(record.id)}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        marginBottom: '16px'
      }}>
        <h2>任务管理</h2>
        <Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={loadTasks}
          >
            刷新
          </Button>
        </Space>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={tasks}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>
    </div>
  );
};

export default Tasks;