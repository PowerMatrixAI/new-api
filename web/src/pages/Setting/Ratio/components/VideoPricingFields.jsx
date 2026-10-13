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

import React, { useState } from 'react';
import { Button, Card, Input, Select, Space } from '@douyinfe/semi-ui';
import { IconDelete, IconPlus } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';

export default function VideoPricingFields({
  model,
  onPriceChange,
  onTiersChange,
}) {
  const { t } = useTranslation();
  const [seconds, setSeconds] = useState('8');
  const changeTier = (index, field, value) => {
    onTiersChange(
      model.videoTiers.map((tier, i) =>
        i === index ? { ...tier, [field]: value } : tier,
      ),
    );
  };
  return (
    <Card
      bodyStyle={{ padding: 16 }}
      style={{ marginBottom: 16, background: 'var(--semi-color-fill-0)' }}
    >
      <div className='font-medium mb-2'>{t('视频每秒价格')}</div>
      <Input
        value={model.videoPrice}
        onChange={onPriceChange}
        placeholder={t('输入每秒生成价格')}
        suffix={t('$/秒')}
      />
      <div className='text-xs text-gray-500 mt-2 mb-4'>
        {t(
          '按生成视频的秒数收费。提交时预扣，完成后按上游返回的实际时长结算；上游未返回时长时保留预扣，失败退款。',
        )}
      </div>
      <div className='font-medium mb-2'>{t('价格档位')}</div>
      <div className='text-xs text-gray-500 mb-3'>
        {t(
          '可按分辨率和是否生成音频设置每秒单价。未匹配时使用基础价格，同时匹配两个条件的档位优先，其次为分辨率档位。',
        )}
      </div>
      {model.videoTiers.map((tier, index) => (
        <div key={index} className='mb-3 p-3 rounded-lg border border-gray-200'>
          <div className='flex items-center justify-between mb-2'>
            <span>
              {t('价格档位')} {index + 1}
            </span>
            <Button
              size='small'
              type='danger'
              theme='borderless'
              icon={<IconDelete />}
              aria-label={t('删除档位')}
              onClick={() =>
                onTiersChange(model.videoTiers.filter((_, i) => i !== index))
              }
            />
          </div>
          <div className='mb-2'>
            <div className='mb-1 text-sm'>{t('输出视频分辨率')}</div>
            <Input
              value={tier.resolution}
              placeholder={t('例如 720p、1080p、4k；留空表示不限')}
              onChange={(value) => changeTier(index, 'resolution', value)}
            />
          </div>
          <div className='mb-2'>
            <div className='mb-1 text-sm'>{t('是否生成音频')}</div>
            <Select
              style={{ width: '100%' }}
              value={tier.audio}
              onChange={(value) => changeTier(index, 'audio', value)}
              optionList={[
                { value: 'any', label: t('不限') },
                { value: 'true', label: t('生成音频') },
                { value: 'false', label: t('不生成音频') },
              ]}
            />
          </div>
          <Input
            value={tier.price}
            placeholder={t('输入每秒生成价格')}
            suffix={t('$/秒')}
            onChange={(value) => changeTier(index, 'price', value)}
          />
        </div>
      ))}
      <Button
        size='small'
        icon={<IconPlus />}
        disabled={model.videoTiers.length >= 64}
        onClick={() =>
          onTiersChange([
            ...model.videoTiers,
            { resolution: '', audio: 'any', price: '' },
          ])
        }
      >
        {t('添加价格档位')}
      </Button>
      <div className='font-medium mt-4 mb-2'>{t('基础价格试算')}</div>
      <Space wrap>
        <Input
          style={{ width: 130 }}
          type='number'
          min={0}
          max={3600}
          value={seconds}
          onChange={setSeconds}
          suffix={t('秒')}
        />
        <span>
          {t('× ${{price}}/秒 = ${{total}}', {
            price: model.videoPrice || '0',
            total: Number.isFinite(Number(seconds) * Number(model.videoPrice))
              ? (Number(seconds) * Number(model.videoPrice)).toFixed(4)
              : '-',
          })}
        </span>
      </Space>
      <div className='text-xs text-gray-500 mt-2'>
        {t('试算未包含分组倍率和价格档位。')}
      </div>
    </Card>
  );
}
