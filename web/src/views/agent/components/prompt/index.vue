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

    <!-- 卡片列表容器 -->
    <div class="cards-wrapper">
      <div 
        class="cards-container" 
        ref="cardsContainer"
        @scroll="handleScroll"
      >
        <div 
          v-for="(card, index) in currentCards" 
          :key="index"
          class="prompt-card"
          @click="handleCardClick(card)"
        >
          <div class="card-title">{{ card.title }}</div>
          <div class="card-description">{{ card.description }}</div>
        </div>
      </div>
      
      <!-- 右侧滚动按钮 -->
      <div 
        class="scroll-button right" 
        v-if="showScrollButton"
        @click="scrollRight"
      >
        <i class="el-icon-arrow-right"></i>
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
      showScrollButton: true,
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
          // 检查是否还有内容可以向右滚动
          const canScroll = container.scrollWidth > container.clientWidth;
          const isAtEnd = container.scrollLeft + container.clientWidth >= container.scrollWidth - 10;
          this.showScrollButton = canScroll && !isAtEnd;
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
  padding: 20px;
  background: #fff;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.prompt-tabs {
  display: flex;
  margin-bottom: 20px;
  gap: 0;
  
  .tab-item {
    padding: 10px 20px;
    cursor: pointer;
    color: #606266;
    font-size: 14px;
    transition: all 0.3s;
    background: #fff;
    border: none;
    white-space: nowrap;
    
    &:hover {
      color: $color;
    }
    
    &.active {
      color: #fff;
      background: $color;
      font-weight: 500;
    }
  }
}

.cards-wrapper {
  position: relative;
  flex: 1;
  overflow: hidden;
}

.cards-container {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 10px 0;
  scroll-behavior: smooth;
  
  // 隐藏滚动条
  &::-webkit-scrollbar {
    display: none;
  }
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.prompt-card {
  flex-shrink: 0;
  width: 280px;
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid transparent;
  
  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    border-color: $color;
    transform: translateY(-2px);
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
  }
}

.scroll-button {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;
  transition: all 0.3s;
  border: 1px solid #e4e7ed;
  
  &:hover {
    background: #f5f7fa;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }
  
  i {
    font-size: 14px;
    color: #606266;
    font-weight: bold;
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
