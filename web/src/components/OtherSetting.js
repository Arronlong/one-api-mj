import React, { useEffect, useState } from 'react';
import { Button, Divider, Form, Grid, Header, Message, Modal } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';
import { marked } from 'marked';

const OtherSetting = () => {
  let [inputs, setInputs] = useState({
    Footer: '',
    Notice: '',
    About: '',
    SystemName: '',
    Logo: '',
    AIChatTitle: '',
    AIChatMainTitle: '',
    AIChatSubTitle: '',
    AIChatModels: '',
    AIChatNoticeShowEnabled: '',
    AIChatNoticeSplashEnabled: '',
    AIChatNoticeTitle: '',
    AIChatNoticeContent: '',
    HomePageContent: ''
  });
  let [loading, setLoading] = useState(false);
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [updateData, setUpdateData] = useState({
    tag_name: '',
    content: ''
  });

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (item.key in inputs) {
          newInputs[item.key] = item.value;
        }
      });
      setInputs(newInputs);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getOptions().then();
  }, []);

  const updateOption = async (key, value) => {
    setLoading(true);
    switch (key) {
      case 'AIChatNoticeShowEnabled':
      case 'AIChatNoticeSplashEnabled':
        value = inputs[key] === 'true' ? 'false' : 'true';
        break;
      default:
        break;
    }
    const res = await API.put('/api/option/', {
      key,
      value
    });
    const { success, message } = res.data;
    if (success) {
      setInputs((inputs) => ({ ...inputs, [key]: value }));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleInputChange = async (e, { name, value }) => {
    if (
        name === 'AIChatNoticeShowEnabled' ||
        name === 'AIChatNoticeSplashEnabled'
    ) {
        await updateOption(name, value);
    } else {
        setInputs((inputs) => ({...inputs, [name]: value}));
    }
  };

  const submitNotice = async () => {
    await updateOption('Notice', inputs.Notice);
  };

  const submitFooter = async () => {
    await updateOption('Footer', inputs.Footer);
  };

  const submitSystemName = async () => {
    await updateOption('SystemName', inputs.SystemName);
  };

  const submitAIChat = async () => {
    await updateOption('AIChatTitle', inputs.AIChatTitle);
    await updateOption('AIChatMainTitle', inputs.AIChatMainTitle);
    await updateOption('AIChatSubTitle', inputs.AIChatSubTitle);
    await updateOption('AIChatModels', inputs.AIChatModels);
    await updateOption('AIChatNoticeTitle', inputs.AIChatNoticeTitle);
    await updateOption('AIChatNoticeContent', inputs.AIChatNoticeContent);
  };

  const submitLogo = async () => {
    await updateOption('Logo', inputs.Logo);
  };

  const submitAbout = async () => {
    await updateOption('About', inputs.About);
  };

  const submitOption = async (key) => {
    await updateOption(key, inputs[key]);
  };

  const openGitHubRelease = () => {
    window.location =
      'https://github.com/songquanpeng/one-api/releases/latest';
  };

  const checkUpdate = async () => {
    const res = await API.get(
      'https://api.github.com/repos/songquanpeng/one-api/releases/latest'
    );
    const { tag_name, body } = res.data;
    if (tag_name === process.env.REACT_APP_VERSION) {
      showSuccess(`已是最新版本：${tag_name}`);
    } else {
      setUpdateData({
        tag_name: tag_name,
        content: marked.parse(body)
      });
      setShowUpdateModal(true);
    }
  };

  return (
    <Grid columns={1}>
      <Grid.Column>
        <Form loading={loading}>
          <Header as='h3'>通用设置</Header>
          <Form.Button onClick={checkUpdate}>检查更新</Form.Button>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='公告'
              placeholder='在此输入新的公告内容，支持 Markdown & HTML 代码'
              value={inputs.Notice}
              name='Notice'
              onChange={handleInputChange}
              style={{ minHeight: 150, fontFamily: 'JetBrains Mono, Consolas' }}
            />
          </Form.Group>
          <Form.Button onClick={submitNotice}>保存公告</Form.Button>
          <Divider />
          <Header as='h3'>个性化设置</Header>
          <Form.Group widths='equal'>
            <Form.Input
              label='系统名称'
              placeholder='在此输入系统名称'
              value={inputs.SystemName}
              name='SystemName'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitSystemName}>设置系统名称</Form.Button>
          <Form.Group widths='equal'>
            <Form.Input
              label='Logo 图片地址'
              placeholder='在此输入 Logo 图片地址'
              value={inputs.Logo}
              name='Logo'
              type='url'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitLogo}>设置 Logo</Form.Button>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='首页内容'
              placeholder='在此输入首页内容，支持 Markdown & HTML 代码，设置后首页的状态信息将不再显示。如果输入的是一个链接，则会使用该链接作为 iframe 的 src 属性，这允许你设置任意网页作为首页。'
              value={inputs.HomePageContent}
              name='HomePageContent'
              onChange={handleInputChange}
              style={{ minHeight: 150, fontFamily: 'JetBrains Mono, Consolas' }}
            />
          </Form.Group>
          <Form.Button onClick={() => submitOption('HomePageContent')}>保存首页内容</Form.Button>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='关于'
              placeholder='在此输入新的关于内容，支持 Markdown & HTML 代码。如果输入的是一个链接，则会使用该链接作为 iframe 的 src 属性，这允许你设置任意网页作为关于页面。'
              value={inputs.About}
              name='About'
              onChange={handleInputChange}
              style={{ minHeight: 150, fontFamily: 'JetBrains Mono, Consolas' }}
            />
          </Form.Group>
          <Form.Button onClick={submitAbout}>保存关于</Form.Button>
          <Message>移除 One API 的版权标识必须首先获得授权，项目维护需要花费大量精力，如果本项目对你有意义，请主动支持本项目。</Message>
          <Form.Group widths='equal'>
            <Form.Input
              label='页脚'
              placeholder='在此输入新的页脚，留空则使用默认页脚，支持 HTML 代码'
              value={inputs.Footer}
              name='Footer'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitFooter}>设置页脚</Form.Button>
          <Divider />
          <Header as='h3'>AIChatWeb设置</Header>
          <Form.Group widths='equal'>
            <Form.Input
              label='AIChatWeb网站Title'
              placeholder='在此输入AIChatWeb网站Title'
              value={inputs.AIChatTitle}
              name='AIChatTitle'
              onChange={handleInputChange}
            />
            <Form.Input
              label='AIChatWeb侧边栏标题'
              placeholder='在此输入AIChatWeb侧边栏标题'
              value={inputs.AIChatMainTitle}
              name='AIChatMainTitle'
              onChange={handleInputChange}
            />
            <Form.Input
              label='AIChatWeb侧边栏副标题'
              placeholder='在此输入AIChat侧边栏副标题'
              value={inputs.AIChatSubTitle}
              name='AIChatSubTitle'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='AIChat模型名'
              placeholder='在此输入AIChat模型名，json数组格式[{name:"gpt-4", contentType:"Text或者Image"}]'
              style={{minHeight: 250, fontFamily: 'JetBrains Mono, Consolas'}}
              value={inputs.AIChatModels}
              name='AIChatModels'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Divider/>
          <Header as='h3'>AIChat公告</Header>
          <Form.Group widths='equal'>
            <Form.Checkbox
                checked={inputs.AIChatNoticeShowEnabled === 'true'}
                label='开启AIChat公告'
                name='AIChatNoticeShowEnabled'
                onChange={handleInputChange}
            />
            <Form.Checkbox
                checked={inputs.AIChatNoticeSplashEnabled === 'true'}
                label='开启AIChat公告展示效果)'
                name='AIChatNoticeSplashEnabled'
                onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.Input
              label='AIChat公告标题'
              placeholder='在此输入AIChat公告标题'
              value={inputs.AIChatNoticeTitle}
              name='AIChatNoticeTitle'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='AIChat公告内容'
              placeholder='在此输入AIChat公告内容'
              style={{minHeight: 250, fontFamily: 'JetBrains Mono, Consolas'}}
              value={inputs.AIChatNoticeContent}
              name='AIChatNoticeContent'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitAIChat}>保存AIChatWeb设置</Form.Button>

        </Form>
      </Grid.Column>
      <Modal
        onClose={() => setShowUpdateModal(false)}
        onOpen={() => setShowUpdateModal(true)}
        open={showUpdateModal}
      >
        <Modal.Header>新版本：{updateData.tag_name}</Modal.Header>
        <Modal.Content>
          <Modal.Description>
            <div dangerouslySetInnerHTML={{ __html: updateData.content }}></div>
          </Modal.Description>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setShowUpdateModal(false)}>关闭</Button>
          <Button
            content='详情'
            onClick={() => {
              setShowUpdateModal(false);
              openGitHubRelease();
            }}
          />
        </Modal.Actions>
      </Modal>
    </Grid>
  );
};

export default OtherSetting;
