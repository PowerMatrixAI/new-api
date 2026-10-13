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

export const serializeVideoPricing = (model, t) => {
  const price = Number(model.videoPrice);
  if (model.videoPrice === '' || !Number.isFinite(price) || price < 0) {
    throw new Error(t('请填写有效的视频每秒价格'));
  }
  if (model.videoTiers.length > 64) {
    throw new Error(t('视频价格最多支持 64 个档位'));
  }
  const seen = new Set();
  const tiers = model.videoTiers.map((tier) => {
    let resolution = tier.resolution.trim().toLowerCase();
    const dimensions = /^(\d+)[x*](\d+)$/.exec(resolution);
    if (dimensions && Number(dimensions[1]) > 0 && Number(dimensions[2]) > 0) {
      resolution = `${Math.min(Number(dimensions[1]), Number(dimensions[2]))}p`;
    }
    const tierPrice = Number(tier.price);
    const key = `${resolution}:${tier.audio}`;
    if (
      (!resolution && tier.audio === 'any') ||
      tier.price === '' ||
      !Number.isFinite(tierPrice) ||
      tierPrice < 0 ||
      seen.has(key)
    ) {
      throw new Error(t('请检查视频价格档位的条件、单价及重复项'));
    }
    seen.add(key);
    return {
      ...(resolution ? { resolution } : {}),
      ...(tier.audio === 'any'
        ? {}
        : { generate_audio: tier.audio === 'true' }),
      price_per_second: tierPrice,
    };
  });
  return { price_per_second: price, ...(tiers.length ? { tiers } : {}) };
};

export const readVideoPricing = (pricing) => ({
  videoPrice: pricing ? String(pricing.price_per_second ?? '') : '',
  videoTiers: (pricing?.tiers || []).map((tier) => ({
    resolution: tier.resolution || '',
    audio: tier.generate_audio == null ? 'any' : String(tier.generate_audio),
    price: String(tier.price_per_second ?? ''),
  })),
});
