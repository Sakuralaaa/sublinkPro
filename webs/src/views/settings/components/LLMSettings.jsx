import { useState, useEffect } from 'react';

// material-ui
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';

// icons
import SmartToyIcon from '@mui/icons-material/SmartToy';
import SaveIcon from '@mui/icons-material/Save';
import SendIcon from '@mui/icons-material/Send';
import VisibilityIcon from '@mui/icons-material/Visibility';
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff';

// project imports
import { getLLMConfig, updateLLMConfig, testLLMConnection } from 'api/settings';

// ==============================|| LLM 设置组件 ||============================== //

export default function LLMSettings({ showMessage, loading, setLoading }) {
  const [form, setForm] = useState({
    apiUrl: '',
    apiKey: '',
    model: 'gpt-3.5-turbo'
  });
  const [showKey, setShowKey] = useState(false);

  useEffect(() => {
    fetchConfig();
  }, []);

  const fetchConfig = async () => {
    try {
      const response = await getLLMConfig();
      if (response.data) {
        setForm({
          apiUrl: response.data.apiUrl || '',
          apiKey: response.data.apiKey || '',
          model: response.data.model || 'gpt-3.5-turbo'
        });
      }
    } catch (error) {
      console.error('获取 LLM 配置失败:', error);
    }
  };

  const handleSave = async () => {
    setLoading(true);
    try {
      await updateLLMConfig(form);
      showMessage('LLM 设置保存成功');
    } catch (error) {
      showMessage('保存失败: ' + (error.message || '未知错误'), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async () => {
    if (!form.apiUrl || !form.apiKey) {
      showMessage('请填写 API URL 和 API Key', 'warning');
      return;
    }

    setLoading(true);
    try {
      await testLLMConnection(form);
      showMessage('LLM API 连接测试成功');
    } catch (error) {
      showMessage('连接测试失败: ' + (error.message || '未知错误'), 'error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader title="LLM API 设置" avatar={<SmartToyIcon color="primary" />} />
      <CardContent>
        <Stack spacing={2.5}>
          <Alert severity="info">
            <Typography variant="body2">
              配置 OpenAI 兼容的 LLM API 接口，可用于智能整理节点和生成订阅规则。支持 OpenAI、DeepSeek、通义千问等兼容接口。
            </Typography>
          </Alert>

          <TextField
            fullWidth
            label="API URL"
            value={form.apiUrl}
            onChange={(e) => setForm({ ...form, apiUrl: e.target.value })}
            placeholder="https://api.openai.com"
            helperText="OpenAI 兼容接口的基础 URL，例如 https://api.openai.com 或 https://api.deepseek.com"
          />

          <TextField
            fullWidth
            label="API Key"
            type={showKey ? 'text' : 'password'}
            value={form.apiKey}
            onChange={(e) => setForm({ ...form, apiKey: e.target.value })}
            placeholder="sk-..."
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton onClick={() => setShowKey(!showKey)} edge="end" size="small">
                    {showKey ? <VisibilityOffIcon /> : <VisibilityIcon />}
                  </IconButton>
                </InputAdornment>
              )
            }}
          />

          <TextField
            fullWidth
            label="模型名称"
            value={form.model}
            onChange={(e) => setForm({ ...form, model: e.target.value })}
            placeholder="gpt-3.5-turbo"
            helperText="使用的模型名称，例如 gpt-4、gpt-3.5-turbo、deepseek-chat 等"
          />

          <Box>
            <Stack direction="row" spacing={2}>
              <Button variant="outlined" color="success" onClick={handleTest} disabled={loading} startIcon={<SendIcon />}>
                测试连接
              </Button>
              <Button variant="contained" onClick={handleSave} disabled={loading} startIcon={<SaveIcon />}>
                保存设置
              </Button>
            </Stack>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}
