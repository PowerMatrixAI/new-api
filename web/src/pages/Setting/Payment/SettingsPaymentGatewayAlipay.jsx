/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState, useRef } from 'react';
import { Banner, Button, Form, Row, Col, Spin } from '@douyinfe/semi-ui';
import { API, removeTrailingSlash, showError, showSuccess } from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsPaymentGatewayAlipay(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    AlipayAppId: '',
    AlipayPrivateKey: '',
    AlipayPublicKey: '',
    AlipaySandbox: false,
    AlipayUnitPrice: 7.3,
    AlipayMinTopUp: 1,
    AlipayAmountOptions: '',
    AlipayAmountDiscount: '{}',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        AlipayAppId: props.options.AlipayAppId || '',
        AlipayPrivateKey: props.options.AlipayPrivateKey || '',
        AlipayPublicKey: props.options.AlipayPublicKey || '',
        AlipaySandbox: props.options.AlipaySandbox === true || props.options.AlipaySandbox === 'true',
        AlipayUnitPrice:
          props.options.AlipayUnitPrice !== undefined
            ? parseFloat(props.options.AlipayUnitPrice)
            : 7.3,
        AlipayMinTopUp:
          props.options.AlipayMinTopUp !== undefined
            ? parseInt(props.options.AlipayMinTopUp)
            : 1,
        AlipayAmountOptions: props.options.AlipayAmountOptions || '',
        AlipayAmountDiscount: props.options.AlipayAmountDiscount || '{}',
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitAlipaySetting = async () => {
    if (!props.options.ServerAddress) {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      const options = [];

      if (inputs.AlipayAppId !== '') {
        options.push({ key: 'AlipayAppId', value: inputs.AlipayAppId });
      }
      if (inputs.AlipayPrivateKey !== '') {
        options.push({ key: 'AlipayPrivateKey', value: inputs.AlipayPrivateKey });
      }
      if (inputs.AlipayPublicKey !== '') {
        options.push({ key: 'AlipayPublicKey', value: inputs.AlipayPublicKey });
      }
      if (inputs.AlipayUnitPrice !== undefined && inputs.AlipayUnitPrice !== null) {
        options.push({ key: 'AlipayUnitPrice', value: inputs.AlipayUnitPrice.toString() });
      }
      if (inputs.AlipayMinTopUp !== undefined && inputs.AlipayMinTopUp !== null) {
        options.push({ key: 'AlipayMinTopUp', value: inputs.AlipayMinTopUp.toString() });
      }
      options.push({ key: 'AlipayAmountOptions', value: inputs.AlipayAmountOptions || '' });
      options.push({ key: 'AlipayAmountDiscount', value: inputs.AlipayAmountDiscount || '{}' });
      if (originInputs['AlipaySandbox'] !== inputs.AlipaySandbox) {
        options.push({ key: 'AlipaySandbox', value: inputs.AlipaySandbox ? 'true' : 'false' });
      }

      const requestQueue = options.map((opt) =>
        API.put('/api/option/', { key: opt.key, value: opt.value }),
      );
      const results = await Promise.all(requestQueue);

      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => showError(res.data.message));
      } else {
        showSuccess(t('更新成功'));
        setOriginInputs({ ...inputs });
        props.refresh?.();
      }
    } catch {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  const serverAddress = props.options?.ServerAddress
    ? removeTrailingSlash(props.options.ServerAddress)
    : t('网站地址');

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={t('支付宝原生支付设置')}>
          <Banner
            type='info'
            description={`${t('异步通知地址（notify_url）填')}：${serverAddress}/api/user/alipay/notify`}
          />
          <Banner
            type='warning'
            description={t('私钥和公钥请填写不含头尾 PEM 标记的纯 Base64 内容（PKCS8 格式）')}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }} style={{ marginTop: 16 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='AlipayAppId'
                label={t('应用 APPID')}
                placeholder={t('支付宝开放平台应用 APPID')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='AlipayUnitPrice'
                precision={2}
                label={t('充值价格（x元/美金）')}
                placeholder={t('例如：7，就是7元/美金')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='AlipayMinTopUp'
                label={t('最低充值美元数量')}
                placeholder={t('例如：1，就是最低充值1$')}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }} style={{ marginTop: 16 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='AlipayPrivateKey'
                label={t('应用私钥（RSA2 PKCS8）')}
                placeholder={t('敏感信息不会发送到前端显示，粘贴后保存即可')}
                autosize={{ minRows: 4 }}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='AlipayPublicKey'
                label={t('支付宝公钥（用于验签）')}
                placeholder={t('支付宝开放平台 → 应用详情 → 支付宝公钥')}
                autosize={{ minRows: 4 }}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }} style={{ marginTop: 16 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayAmountOptions'
                label={t('充值额度选项')}
                placeholder={t('逗号分隔的美元金额，例如 10,20,50,100,200；为空则自动生成')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayAmountDiscount'
                label={t('充值折扣配置（JSON）')}
                placeholder='{"100":0.9,"500":0.85}'
                extraText={t('键为充值美金数量，值为折扣率；例如 0.9 表示 9 折')}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }} style={{ marginTop: 16 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='AlipaySandbox'
                size='default'
                checkedText='｜'
                uncheckedText='〇'
                label={t('使用沙箱环境')}
              />
            </Col>
          </Row>
          <Button onClick={submitAlipaySetting} style={{ marginTop: 16 }}>
            {t('更新支付宝设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
