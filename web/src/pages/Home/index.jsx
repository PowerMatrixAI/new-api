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

import React, { useContext, useEffect, useState, useCallback } from 'react';
import { API, showError } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import NoticeModal from '../../components/layout/NoticeModal';
import { Link } from 'react-router-dom';
import './PowerMatrixHome.css';
import logoImg from './assets/logo.png';
import qrcodeImg from './assets/qrcode.png';

// ============================================================
// PowerMatrix Landing Page — React Component
// ============================================================

const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const actualTheme = useActualTheme();
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const isMobile = useIsMobile();
  const isDemoSiteMode = statusState?.status?.demo_site_enabled || false;

  // Solution tabs
  const [activeTab, setActiveTab] = useState('agent');
  // FAQ open state
  const [openFaq, setOpenFaq] = useState(null);
  // Navbar scroll
  const [navScrolled, setNavScrolled] = useState(false);

  const toggleFaq = useCallback((idx) => {
    setOpenFaq((prev) => (prev === idx ? null : idx));
  }, []);

  useEffect(() => {
    const handleScroll = () => setNavScrolled(window.scrollY > 50);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);
      if (data.startsWith('https://')) {
        const iframe = document.querySelector('iframe');
        if (iframe) {
          iframe.onload = () => {
            iframe.contentWindow.postMessage({ themeMode: actualTheme }, '*');
            iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
          };
        }
      }
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };
    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent().then();
  }, []);

  // ============================================================
  // If admin set custom homepage content, show that instead
  // ============================================================
  if (homePageContentLoaded && homePageContent !== '') {
    return (
      <div className='w-full overflow-x-hidden'>
        <NoticeModal
          visible={noticeVisible}
          onClose={() => setNoticeVisible(false)}
          isMobile={isMobile}
        />
        <div className='overflow-x-hidden w-full'>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              className='w-full h-screen border-none'
            />
          ) : (
            <div
              className='mt-[60px]'
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            />
          )}
        </div>
      </div>
    );
  }

  // ============================================================
  // PowerMatrix Landing Page
  // ============================================================
  const solutions = {
    agent: {
      title: 'AI Agent 工作站底座',
      desc: '为多智能体协同提供稳定且高频的算力燃料。从任务规划、环境交互到智能体间的通信验证，完美支撑中小型企业的自动化运营需求。',
      tags: ['多智能体协同', '高频调用', '自动化运营', '任务规划'],
      nodes: [
        { color: 'green', label: '任务规划 Agent', sub: 'GPT-4o' },
        { color: 'blue', label: 'PowerMatrix 路由层', sub: '智能调度' },
        { color: 'yellow', label: '执行 Agent', sub: 'Claude / Gemini' },
      ],
    },
    content: {
      title: '自动化内容与营销引擎',
      desc: '结合企业私有知识库，海量生成各类营销文案。提供流畅的长文本理解与输出能力，轻松应对私域客服意图识别与售后工单处理。',
      tags: ['内容批量生成', '私域客服', '长文本处理', '知识库融合'],
      nodes: [
        { color: 'blue', label: '企业知识库', sub: 'RAG 向量检索' },
        { color: 'green', label: '营销文案生成', sub: 'Claude 3.5 Sonnet' },
        { color: 'yellow', label: '多渠道分发', sub: '微博 / 公众号 / 抖音' },
      ],
    },
    data: {
      title: '复杂商业数据结构化',
      desc: '从海量冗杂的行业新闻、长篇报告中提取核心变量。利用强大的逻辑推理能力，秒级输出严格的 JSON 格式，为量化分析与商业决策系统提供纯净数据。',
      tags: ['数据提取', 'JSON 输出', '量化分析', '逻辑推理'],
      nodes: [
        { color: 'yellow', label: '原始数据输入', sub: '行业报告 / 新闻流' },
        { color: 'blue', label: '推理提取层', sub: 'Gemini 1.5 Pro' },
        { color: 'green', label: '结构化输出', sub: '严格 JSON Schema' },
      ],
    },
  };

  const pricingPlans = [
    {
      tier: '基础尝鲜',
      price: '¥9',
      unit: '/月',
      target: '适合学生与轻度尝鲜用户',
      features: [
        '50 万算力积分（月底清零）',
        '极速访问国内顶尖大模型',
        '保障日常从容对话',
        '满足轻度问答交互',
      ],
      btn: '立即订阅',
    },
    {
      tier: '标准生产',
      price: '¥29',
      unit: '/月',
      target: '适合自媒体与内容创作者',
      features: [
        '180 万算力积分',
        '进阶长文本与多模态模型',
        '流畅连续交互',
        '无缝应对高频写作需求',
      ],
      btn: '立即订阅',
    },
    {
      tier: '极客专业',
      price: '¥149',
      unit: '/月',
      target: '适合独立开发者与极客团队',
      features: [
        '1000 万算力积分',
        '全量解锁海外旗舰大模型库',
        '专属极速请求通道',
        '无惧复杂自动化脚本调用',
      ],
      btn: '立即订阅',
      featured: true,
    },
    {
      tier: '企业商用',
      price: '¥599',
      unit: '/月起',
      target: '适合中小型企业与 SaaS 平台',
      features: [
        '超大额度定制积分池',
        '团队多级子账号管理',
        '定制化速率上限',
        '正规企业对公发票',
      ],
      btn: '联系销售',
    },
  ];

  const faqs = [
    {
      q: '接入 PowerMatrix 需要修改原有代码吗？',
      a: '几乎不需要。我们的接口 100% 兼容 OpenAI 格式，您只需将代码中的 Base URL 和 API Key 替换为我们的，即可瞬间完成迁移，所有现有业务逻辑零修改。',
    },
    {
      q: '流量加油包和套餐内的积分有什么区别？',
      a: '套餐内的保底积分在每月账单日清零，确保基础的日常消耗；而单独购买的「流量加油包」积分永久有效，优先抵扣超出部分，给您的业务带来最大弹性和安全感。',
    },
    {
      q: '我们的数据隐私如何保障？',
      a: '我们仅作为纯粹的 API 路由网关，系统配置了严格的数据无痕流转机制，绝不缓存、记录或使用您的任何业务对话数据，您的商业机密始终属于您自己。',
    },
    {
      q: '可以开具增值税发票吗？',
      a: '完全支持。我们为国内企业用户提供正规的增值税普通/专用发票，轻松解决团队财务报销难题，欢迎联系企业客服进行对公认证。',
    },
  ];

  const activeSol = solutions[activeTab];

  return (
    <div className='pm-landing'>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />

      {/* ===== NAVBAR ===== */}
      <nav className={`pm-navbar ${navScrolled ? 'scrolled' : ''}`}>
        <div className='pm-navbar-inner'>
          <div className='pm-nav-left'>
            <a href='#' className='pm-nav-logo'>
              <img src={logoImg} alt='PowerMatrix Logo' className='pm-logo-img' />
            </a>
            <ul className='pm-nav-links'>
              <li><Link to='/'>首页</Link></li>
              <li><Link to='/console'>控制台</Link></li>
              <li><Link to='/pricing'>模型广场</Link></li>
            </ul>
          </div>
          <div className='pm-nav-actions'>
            <div className='pm-lang-toggle'>
              <button
                className={`pm-lang-btn ${i18n.language.startsWith('zh') ? 'active' : ''}`}
                onClick={() => i18n.changeLanguage('zh-CN')}
              >中文</button>
              <span className='pm-lang-sep'>|</span>
              <button
                className={`pm-lang-btn ${i18n.language === 'en' ? 'active' : ''}`}
                onClick={() => i18n.changeLanguage('en')}
              >EN</button>
            </div>
            <Link to='/login' className='pm-btn-ghost'>登录 / 注册</Link>
            <Link to='/console' className='pm-btn-nav-primary'>获取 API 密钥</Link>
          </div>
        </div>
      </nav>

      {/* ===== HERO ===== */}
      <section className='pm-hero' id='pm-hero'>
        <div className='pm-hero-bg'>
          <div className='pm-hero-grid' />
          <div className='pm-hero-glow' />
          <div className='pm-hero-glow2' />
        </div>
        <div className='pm-hero-content'>
          <div className='pm-hero-text'>
            <div className='pm-hero-badge'>
              <span className='pm-hero-badge-dot' />
              <span>企业级 API 基础设施 · 2026</span>
            </div>
            <h1>
              PowerMatrix<br />
              <span className='pm-highlight'>算力矩阵</span><br />
              大模型接口枢纽
            </h1>
            <p className='pm-hero-sub'>
              极简接入全球顶尖 AI 算力。一次配置，全量调用国内外前沿大模型。告别繁琐计费与网络阻力，专注构建你的核心 AI 业务。
            </p>
            <div className='pm-hero-ctas'>
              <Link to='/console'>
                <button className='pm-btn-cta-primary'>
                  <span>免费获取密钥</span>
                  <span className='pm-btn-cta-hint'>⚡ 1分钟极速接入</span>
                </button>
              </Link>
              <button
                className='pm-btn-cta-secondary'
                onClick={() => {
                  const docsLink = statusState?.status?.docs_link;
                  if (docsLink) window.open(docsLink, '_blank');
                }}
              >
                查看开发文档 →
              </button>
            </div>
          </div>

          {/* Code Block */}
          <div className='pm-hero-code'>
            <div className='pm-code-header'>
              <span className='pm-code-dot pm-code-dot-r' />
              <span className='pm-code-dot pm-code-dot-y' />
              <span className='pm-code-dot pm-code-dot-g' />
              <span className='pm-code-title'>bash · 极速接入示例</span>
            </div>
            <pre>
              <code>
                <span className='pm-t-comment'># 替换 Base URL 即可无缝迁移</span>{'\n'}
                <span className='pm-t-cmd'>curl</span> https://api.powermatrix.tech/v1/chat/completions \{'\n'}
                {'  '}<span className='pm-t-flag'>-H</span> <span className='pm-t-string'>"Authorization: Bearer YOUR_POWER_KEY"</span> \{'\n'}
                {'  '}<span className='pm-t-flag'>-H</span> <span className='pm-t-string'>"Content-Type: application/json"</span> \{'\n'}
                {'  '}<span className='pm-t-flag'>-d</span> <span className='pm-t-string'>{"'{"}</span>{'\n'}
                {'    '}<span className='pm-t-key'>"model"</span>: <span className='pm-t-val'>"gpt-4o"</span>,   <span className='pm-t-comment'># 数十款顶尖模型随意切换</span>{'\n'}
                {'    '}<span className='pm-t-key'>"messages"</span>: {'[{'}{'\n'}
                {'      '}<span className='pm-t-key'>"role"</span>: <span className='pm-t-val'>"user"</span>,{'\n'}
                {'      '}<span className='pm-t-key'>"content"</span>: <span className='pm-t-val'>"启动智能体引擎"</span>{'\n'}
                {'    '}{'}'}{']'}{'\n'}
                {'  '}<span className='pm-t-string'>{"'}"}</span>
              </code>
            </pre>
          </div>
        </div>
      </section>

      {/* ===== FEATURES ===== */}
      <section className='pm-section' id='pm-features'>
        <div className='pm-section-inner'>
          <div className='pm-section-label'>核心优势</div>
          <h2 className='pm-section-title'>为什么选择 PowerMatrix？</h2>
          <p className='pm-section-sub'>
            我们将复杂的全球大模型接入问题化繁为简，让您专注于创造真正的业务价值。
          </p>
          <div className='pm-features-grid'>
            {[
              { icon: '⚡', title: '全局智能路由', desc: '底层动态融合多重专线与优质节点。从简单的内容生成到复杂的自动化指令，请求将被智能分配，确保全天候极速响应。' },
              { icon: '💰', title: '计费透明可控', desc: '采用统一的算力积分系统，支持微信、支付宝及对公账户。消除跨国支付壁垒与汇率磨损，将 API 调用成本降至极致。' },
              { icon: '🛡️', title: '企业级稳定性', desc: '多渠道自动容灾与回退机制，为关键业务提供强力兜底。支持高并发请求，保障您的生产环境坚如磐石。' },
              { icon: '🔌', title: '100% 协议兼容', desc: '完全兼容 OpenAI 标准接口。无需重构已有代码，一行代码即可无缝接入各类主流智能体框架与应用底座。' },
            ].map((f, i) => (
              <div className='pm-feature-card' key={i}>
                <div className='pm-feature-icon'>{f.icon}</div>
                <div className='pm-feature-title'>{f.title}</div>
                <p className='pm-feature-desc'>{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <div className='pm-divider' />

      {/* ===== SOLUTIONS ===== */}
      <section className='pm-section pm-solutions-bg' id='pm-solutions'>
        <div className='pm-section-inner'>
          <div className='pm-section-label'>业务落地场景</div>
          <h2 className='pm-section-title'>支撑复杂商业需求</h2>
          <p className='pm-section-sub'>
            PowerMatrix 不止是 API 中转，更是您 AI 业务的坚实算力底座。
          </p>
          <div className='pm-tabs'>
            {[
              { key: 'agent', label: 'AI Agent 底座' },
              { key: 'content', label: '自动化内容引擎' },
              { key: 'data', label: '商业数据结构化' },
            ].map((tab) => (
              <button
                key={tab.key}
                className={`pm-tab-btn ${activeTab === tab.key ? 'active' : ''}`}
                onClick={() => setActiveTab(tab.key)}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <div className='pm-tab-content active'>
            <div>
              <h3 className='pm-solution-title'>{activeSol.title}</h3>
              <p className='pm-solution-desc'>{activeSol.desc}</p>
              <div className='pm-solution-tags'>
                {activeSol.tags.map((tag, i) => (
                  <span className='pm-tag' key={i}>{tag}</span>
                ))}
              </div>
            </div>
            <div className='pm-solution-visual'>
              {activeSol.nodes.map((node, i) => (
                <React.Fragment key={i}>
                  {i > 0 && <div className='pm-sol-arrow'>↕</div>}
                  <div className='pm-sol-node'>
                    <span className={`pm-sol-dot pm-sol-dot-${node.color}`} />
                    <div className='pm-sol-line'>
                      <strong>{node.label}</strong> — {node.sub}
                    </div>
                  </div>
                </React.Fragment>
              ))}
            </div>
          </div>
        </div>
      </section>

      <div className='pm-divider' />

      {/* ===== PRICING ===== */}
      <section className='pm-section' id='pm-pricing'>
        <div className='pm-section-inner'>
          <div className='pm-section-label'>灵活订阅制计费</div>
          <h2 className='pm-section-title'>选择适合你的方案</h2>
          <p className='pm-section-sub'>
            从个人尝鲜到企业商用，全阶段覆盖。随时升降级，积分灵活充值。
          </p>
          <div className='pm-pricing-grid'>
            {pricingPlans.map((plan, i) => (
              <div
                className={`pm-pricing-card ${plan.featured ? 'pm-featured' : ''}`}
                key={i}
              >
                {plan.featured && (
                  <div className='pm-badge-popular'>🔥 最受欢迎</div>
                )}
                <div className='pm-price-tier'>{plan.tier}</div>
                <div className='pm-price-amount'>
                  {plan.price}<span>{plan.unit}</span>
                </div>
                <div className='pm-price-target'>{plan.target}</div>
                <ul className='pm-price-features'>
                  {plan.features.map((f, j) => (
                    <li key={j}>
                      <span className='pm-check'>✓</span>
                      <span>{f}</span>
                    </li>
                  ))}
                </ul>
                <Link to='/console/topup'>
                  <button
                    className={`pm-btn-price ${plan.featured ? 'pm-featured' : ''}`}
                  >
                    {plan.btn}
                  </button>
                </Link>
              </div>
            ))}
          </div>
          <div className='pm-pricing-note'>
            💡 积分不够用？所有套餐均可随时购买{' '}
            <strong>永久有效的流量加油包</strong>
            ，随充随用，永不断连，给您的业务带来最大弹性与安全感。
          </div>
        </div>
      </section>

      <div className='pm-divider' />

      {/* ===== FAQ ===== */}
      <section className='pm-section' id='pm-faq'>
        <div className='pm-section-inner'>
          <div className='pm-section-label' style={{ textAlign: 'center' }}>常见问题</div>
          <h2 className='pm-section-title' style={{ textAlign: 'center' }}>
            接入前的最后顾虑
          </h2>
          <p
            className='pm-section-sub'
            style={{ textAlign: 'center', margin: '0 auto 3rem' }}
          >
            我们为您一一解答
          </p>
          <div className='pm-faq-list'>
            {faqs.map((faq, i) => (
              <div
                className={`pm-faq-item ${openFaq === i ? 'open' : ''}`}
                key={i}
              >
                <button className='pm-faq-q' onClick={() => toggleFaq(i)}>
                  <span>{faq.q}</span>
                  <span className='pm-faq-arrow'>▾</span>
                </button>
                <div className='pm-faq-a'>
                  <div className='pm-faq-a-inner'>{faq.a}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ===== FOOTER ===== */}
      <footer className='pm-footer'>
        <div className='pm-footer-inner'>
          <div className='pm-footer-top'>
            <div className='pm-footer-brand'>
              <a href='#' className='pm-footer-logo-link'>
                <img
                  src={logoImg}
                  alt='PowerMatrix Logo'
                  className='pm-footer-logo-img'
                />
              </a>
              <p>
                连接智能，驱动未来。<br />
                企业级大模型 API 接入基础设施。
              </p>
              <div className='pm-footer-qr-wrap'>
                <div className='pm-footer-qr-trigger'>
                  <span>💬</span>
                  <span className='pm-footer-qr-label'>扫码联系企业微信</span>
                  <span className='pm-footer-qr-arrow'>›</span>
                </div>
                <div className='pm-footer-qr-popup'>
                  <img
                    src={qrcodeImg}
                    alt='Enterprise WeChat QR Code'
                    className='pm-footer-qr-img'
                  />
                  <p className='pm-footer-qr-tip'>微信扫码，联系商务顾问</p>
                </div>
              </div>
            </div>
          </div>
          <div className='pm-footer-bottom'>
            <p>© 2026 杭州哈希降临科技有限公司 版权所有。</p>
            <p style={{ color: 'var(--pm-gray-500)', fontSize: '0.8rem' }}>
              api.powermatrix.tech
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Home;
