<template>
  <div class="prompt-template-container">
    <!-- 标签页 -->
    <div class="prompt-tabs">
      <div 
        class="tab-item" 
        :class="{ active: activeTab === 'recommended' }"
        @click="activeTab = 'recommended'"
      >
        {{ $t('agent.promptTemplate.recommended') }}
      </div>
      <div 
        class="tab-item" 
        :class="{ active: activeTab === 'personal' }"
        @click="activeTab = 'personal'"
      >
        {{ $t('agent.promptTemplate.personal')}}
      </div>
    </div>

    <div class="cards-wrapper">
      <!-- 空状态展示 -->
      <div v-if="showEmptyState" class="empty-state">
        <div class="empty-icon">
          <svg width="64" height="64" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
            <!-- 三条短线 -->
            <path d="M20 24H44" stroke="#D9D9D9" stroke-width="2" stroke-linecap="round"/>
            <path d="M20 32H44" stroke="#D9D9D9" stroke-width="2" stroke-linecap="round"/>
            <path d="M20 40H44" stroke="#D9D9D9" stroke-width="2" stroke-linecap="round"/>
            <!-- 盒子/托盘 -->
            <path d="M16 12H48C49.1046 12 50 12.8954 50 14V50C50 51.1046 49.1046 52 48 52H16C14.8954 52 14 51.1046 14 50V14C14 12.8954 14.8954 12 16 12Z" stroke="#D9D9D9" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M14 18H50" stroke="#D9D9D9" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <div class="empty-text">当前空间暂无可用的提示词资源</div>
      </div>
      
      <!-- 卡片列表 -->
      <div 
        v-else
        class="cards-container" 
        ref="cardsContainer"
        @scroll="handleScroll"
      >
        <div 
          class="scroll-button left" 
          v-if="showLeftButton"
          @click="scrollLeft"
        >
          <i class="el-icon-arrow-left"></i>
        </div>
        <div 
          v-for="(card, index) in currentCards" 
          :key="index"
          class="prompt-card"
          @click="handleCardClick(card)"
        >
          <div class="card-title">{{ card.title }}</div>
          <div class="card-description">{{ card.description }}</div>
        </div>
        
        <!-- 全部卡片 -->
        <div 
          v-if="activeTab === 'recommended'"
          class="prompt-card all-card"
          @click="handleAllClick"
        >
          <div class="all-card-content">
            <div class="all-card-text">全部</div>
          </div>
        </div>
        
        <div 
          class="scroll-button right" 
          v-if="showRightButton"
          @click="scrollRight"
        >
          <i class="el-icon-arrow-right"></i>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PromptTemplate',
  data() {
    return {
      activeTab: 'recommended',
      showLeftButton: false,
      showRightButton: true,
      recommendedCards: [
        {
          title: '通用结构',
          description: '适用于多种场景的提示词结构,可以根据具体需求增删对应模块',
          type: 'general'
        },
        {
          title: '任务执行',
          description: '适用于有明确的工作步骤的任务执行场景,通过明确每一步骤的工作要求来指导模型完成复杂任务',
          type: 'task'
        },
        {
          title: '角色扮演',
          description: '适用于聊天陪伴、互动娱乐场景,可帮助模型轻松塑造个性化的人物角色,提升对话趣味性',
          type: 'roleplay'
        },
        {
          title: '技能调优',
          description: '适用于需要获取特定技能的场景,通过结构化提示词优化模型在特定领域的表现',
          type: 'skill'
        }
      ],
      personalCards: []
    }
  },
  computed: {
    currentCards() {
      return this.activeTab === 'recommended' ? this.recommendedCards : this.personalCards;
    },
    showEmptyState() {
      return this.activeTab === 'personal' && (!this.personalCards || this.personalCards.length === 0);
    }
  },
  mounted() {
    this.checkScrollButton();
    // 监听窗口大小变化
    window.addEventListener('resize', this.checkScrollButton);
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkScrollButton);
  },
  methods: {
    handleCardClick(card) {
      // 触发卡片点击事件，可以传递给父组件
      this.$emit('card-click', card);
    },
    handleAllClick() {
      // 触发全部卡片点击事件
      this.$emit('all-click');
    },
    scrollLeft() {
      const container = this.$refs.cardsContainer;
      if (container) {
        const scrollAmount = 300; // 每次滚动300px
        container.scrollBy({
          left: -scrollAmount,
          behavior: 'smooth'
        });
      }
    },
    scrollRight() {
      const container = this.$refs.cardsContainer;
      if (container) {
        const scrollAmount = 300; // 每次滚动300px
        container.scrollBy({
          left: scrollAmount,
          behavior: 'smooth'
        });
      }
    },
    handleScroll() {
      this.checkScrollButton();
    },
    checkScrollButton() {
      this.$nextTick(() => {
        const container = this.$refs.cardsContainer;
        if (container) {
          const canScroll = container.scrollWidth > container.clientWidth;
          const scrollLeft = container.scrollLeft;
          const scrollWidth = container.scrollWidth;
          const clientWidth = container.clientWidth;
          
          // 判断是否在开始位置（允许10px的误差）
          const isAtStart = scrollLeft <= 10;
          // 判断是否在结束位置（允许10px的误差）
          const isAtEnd = scrollLeft + clientWidth >= scrollWidth - 10;
          
          // 如果无法滚动，都不显示按钮
          if (!canScroll) {
            this.showLeftButton = false;
            this.showRightButton = false;
          } else {
            // 在开始位置，只显示右侧按钮
            if (isAtStart) {
              this.showLeftButton = false;
              this.showRightButton = true;
            } 
            // 在结束位置，只显示左侧按钮
            else if (isAtEnd) {
              this.showLeftButton = true;
              this.showRightButton = false;
            } 
            // 在中间位置，两个按钮都显示
            else {
              this.showLeftButton = true;
              this.showRightButton = true;
            }
          }
        }
      });
    }
  },
  watch: {
    activeTab() {
      this.$nextTick(() => {
        this.checkScrollButton();
      });
    }
  }
}
</script>

<style lang="scss" scoped>
.prompt-template-container {
  position:absolute;
  bottom:0;
  left:0;
  right:0;
  width: 100%;
  padding: 10px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.prompt-tabs {
  display: flex;
  gap: 10;
  .tab-item {
    padding: 3px 8px;
    cursor: pointer;
    color: #303133;
    font-size: 14px;
    transition: all 0.3s;
    border: none;
    white-space: nowrap;
    border-radius: 4px 4px 0 0;
    
    &:hover {
      color: $color;
    }
    
    &.active {
      color: $color;
      background: #E0E7FF;
      font-weight: 500;
    }
  }
}

.cards-wrapper {
  position: relative;
  flex: 1;
  overflow: hidden;
  width: 100%;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 10px 0;
  
  .empty-icon {
    margin-bottom: 16px;
    opacity: 0.6;
    
    svg {
      display: block;
    }
  }
  
  .empty-text {
    font-size: 14px;
    color: #606266;
    text-align: center;
    line-height: 1.5;
  }
}

.cards-container {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 10px 0;
  scroll-behavior: smooth;
  position: relative;
  align-items: stretch;
  width: 100%;
  max-width: 100%;
  
  // 隐藏滚动条
  &::-webkit-scrollbar {
    display: none;
  }
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.prompt-card {
  flex-shrink: 0;
  width: 200px;
  background: #fff;
  border-radius: 8px;
  padding: 10px;
  box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid transparent;
  
  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    border-color: $color;
    transform: translateY(-2px);
  }
  
  &.all-card {
    background: #fff;
    border: 1px solid transparent;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    
    &:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      border-color: $color;
      transform: translateY(-2px);
    }
    
    .all-card-content {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      width: 100%;
      height: 100%;
      padding: 20px;
      
      .all-card-text {
        font-size: 16px;
        font-weight: 600;
        color: #303133;
      }
    }
  }
  
  .card-title {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 12px;
    line-height: 1.5;
  }
  
  .card-description {
    font-size: 13px;
    color: #606266;
    line-height: 1.6;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
    word-break: break-word;
    max-height: calc(1.6em * 3);
  }
}

.scroll-button {
  position: sticky;
  top: auto;
  align-self: center;
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  min-width: 32px;
  min-height: 32px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 5;
  transition: all 0.3s;
  border: 1px solid #e4e7ed;
  
  &.left {
    left: 0;
    margin-right: 8px;
  }
  
  &.right {
    right: 0;
    margin-left: 8px;
  }
  
  &:hover {
    background: #f5f7fa;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }
  
  i {
    font-size: 14px;
    color: #606266;
    font-weight: bold;
    line-height: 1;
  }
  
  &:hover i {
    color: $color;
  }
}

// 响应式设计
@media (max-width: 768px) {
  .prompt-card {
    min-width: 240px;
    max-width: 240px;
  }
}
</style>
