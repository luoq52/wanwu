<template>
  <div class="graph-map-container">
    <div class="graph-toolbar">
      <el-tooltip content="放大" placement="top">
        <el-button 
          icon="el-icon-zoom-in" 
          circle 
          size="mini" 
          @click="zoomIn"
        ></el-button>
      </el-tooltip>
      <el-tooltip content="缩小" placement="top">
        <el-button 
          icon="el-icon-zoom-out" 
          circle 
          size="mini" 
          @click="zoomOut"
        ></el-button>
      </el-tooltip>
      <el-tooltip content="适应画布" placement="top">
        <el-button 
          icon="el-icon-full-screen" 
          circle 
          size="mini" 
          @click="fitView"
        ></el-button>
      </el-tooltip>
      <el-tooltip content="实际大小" placement="top">
        <el-button 
          icon="el-icon-refresh-left" 
          circle 
          size="mini" 
          @click="resetZoom"
        ></el-button>
      </el-tooltip>
      <el-divider direction="vertical"></el-divider>
      <el-tooltip content="刷新数据" placement="top">
        <el-button 
          icon="el-icon-refresh" 
          circle 
          size="mini" 
          :loading="loading"
          @click="refreshData"
        ></el-button>
      </el-tooltip>
    </div>
    <div ref="graphContainer" class="graph-container" v-loading="loading"></div>
  </div>
</template>

<script>
import G6 from '@antv/g6'

export default {
  name: 'GraphMap',
  props: {
    // 图谱数据
    data: {
      type: Object,
      default: function() {
        return {
          nodes: [],
          edges: []
        }
      }
    },
    // 是否自动适应画布
    autoFit: {
      type: Boolean,
      default: true
    },
    // 布局类型
    layout: {
      type: String,
      default: 'force' // force, dagre, circular, radial, etc.
    },
    // 节点样式配置
    nodeStyle: {
      type: Object,
      default: function() {
        return {}
      }
    },
    // 边样式配置
    edgeStyle: {
      type: Object,
      default: function() {
        return {}
      }
    },
    // 是否使用内置假数据
    useMock: {
      type: Boolean,
      default: false
    },
    // 假数据节点数量
    mockNodes: {
      type: Number,
      default: 200
    },
    // 假数据边数量
    mockEdges: {
      type: Number,
      default: 300
    },
    // 性能优化配置
    performance: {
      type: Object,
      default: function() {
        return {
          largeThreshold: 500,
          hideLabelZoom: 0.8,
          enableWorkerLayout: true,
          simplifyEdgeOnLarge: true,
          disableAnimateOnLarge: true
        }
      }
    }
  },
  data() {
    return {
      graph: null,
      loading: false,
      zoom: 1,
      lastLabelSwitchRAF: 0
    }
  },
  watch: {
    data: {
      handler(newData) {
        if (this.graph && newData) {
          this.updateGraphData(newData)
        }
      },
      deep: true,
      immediate: false
    }
  },
  mounted() {
    this.initGraph()
    if (this.useMock && (!this.data || (!this.data.nodes && !this.data.edges) || (this.data.nodes || []).length === 0)) {
      const mock = this.generateMockData(this.mockNodes, this.mockEdges)
      this.updateGraphData(mock)
    }
  },
  beforeDestroy() {
    if (this.graph) {
      this.graph.destroy()
      this.graph = null
    }
    window.removeEventListener('resize', this.handleResize)
  },
  methods: {
    // 初始化图谱
    initGraph() {
      if (!this.$refs.graphContainer) {
        return
      }

      if (!G6) {
        console.error('G6 is not available')
        return
      }

      const container = this.$refs.graphContainer
      const width = container.clientWidth || 800
      const height = container.clientHeight || 600

      // 注册自定义节点
      this.registerCustomNodes()

      // 获取 Graph 构造函数（兼容不同的导入方式）
      const Graph = G6.Graph || G6

      // 创建图谱实例
      const graphConfig = {
        container: container,
        width: width,
        height: height,
        animate: false,
        modes: {
          default: [
            'drag-canvas',
            'zoom-canvas',
            'drag-node',
            'click-select',
            'brush-select'
          ]
        },
        layout: {
          type: this.layout,
          preventOverlap: true,
          nodeSize: 50,
          nodeSpacing: 30,
          ...this.getLayoutConfig()
        },
        defaultNode: {
          type: 'circle',
          size: 50,
          labelCfg: {
            style: {
              fill: '#333',
              fontSize: 12
            }
          },
          style: {
            fill: '#C6E5FF',
            stroke: '#5B8FF9',
            lineWidth: 2
          },
          ...this.nodeStyle
        },
        defaultEdge: {
          type: 'line',
          style: {
            stroke: '#A3B1BF',
            lineWidth: 2,
            endArrow: true
          },
          labelCfg: {
            autoRotate: true,
            style: {
              fill: '#666',
              fontSize: 10,
              background: {
                fill: '#fff',
                padding: [2, 2, 2, 2],
                radius: 2
              }
            }
          },
          ...this.edgeStyle
        },
        nodeStateStyles: {
          hover: {
            fill: '#91d5ff',
            stroke: '#1890ff',
            lineWidth: 3
          },
          selected: {
            fill: '#91d5ff',
            stroke: '#1890ff',
            lineWidth: 3
          }
        },
        edgeStateStyles: {
          hover: {
            stroke: '#1890ff',
            lineWidth: 3
          },
          selected: {
            stroke: '#1890ff',
            lineWidth: 3
          }
        }
      }

      this.graph = new Graph(graphConfig)

      // 监听事件
      this.bindEvents()

      // 加载数据
      if (this.data && (this.data.nodes || this.data.edges)) {
        const prepared = this.prepareDataForPerformance(this.data)
        this.graph.data(prepared)
        this.graph.render()
        
        if (this.autoFit) {
          this.fitView()
        }
      }

      // 监听窗口大小变化
      window.addEventListener('resize', this.handleResize)
    },

    // 注册自定义节点
    registerCustomNodes() {
      // 可以在这里注册自定义节点类型
      // 例如：G6.registerNode('custom-node', {...})
    },

    // 获取布局配置
    getLayoutConfig() {
      const configs = {
        force: {
          preventOverlap: true,
          nodeSize: 50,
          linkDistance: 100,
          nodeStrength: -50,
          edgeStrength: 0.2
        },
        dagre: {
          rankdir: 'TB',
          nodesep: 50,
          ranksep: 50
        },
        circular: {
          radius: 200,
          startRadius: 10,
          endRadius: 300
        },
        radial: {
          unitRadius: 100,
          nodeSize: 50
        }
      }
      return configs[this.layout] || configs.force
    },

    // 绑定事件
    bindEvents() {
      if (!this.graph) return

      // 节点点击事件
      this.graph.on('node:click', (e) => {
        const node = e.item
        this.$emit('node-click', node.getModel())
      })

      // 节点鼠标进入
      this.graph.on('node:mouseenter', (e) => {
        const node = e.item
        this.graph.setItemState(node, 'hover', true)
      })

      // 节点鼠标离开
      this.graph.on('node:mouseleave', (e) => {
        const node = e.item
        this.graph.setItemState(node, 'hover', false)
      })

      // 边点击事件
      this.graph.on('edge:click', (e) => {
        const edge = e.item
        this.$emit('edge-click', edge.getModel())
      })

      // 画布点击事件
      this.graph.on('canvas:click', () => {
        this.graph.getNodes().forEach(node => {
          this.graph.setItemState(node, 'selected', false)
        })
        this.graph.getEdges().forEach(edge => {
          this.graph.setItemState(edge, 'selected', false)
        })
      })

      // 缩放事件
      this.graph.on('viewportchange', () => {
        if (this.graph) {
          const zoom = this.graph.getZoom()
          this.zoom = zoom
          this.$emit('zoom-change', zoom)
          this.toggleLabelsByZoom()
        }
      })
    },

    // 更新图谱数据
    updateGraphData(data) {
      if (!this.graph) return

      const prepared = this.prepareDataForPerformance(data)
      this.graph.data(prepared)
      this.graph.render()
      
      if (this.autoFit) {
        this.$nextTick(() => {
          this.fitView()
        })
      }
    },

    // 根据缩放控制标签显隐
    toggleLabelsByZoom() {
      if (!this.graph) return
      const threshold = (this.performance && this.performance.hideLabelZoom) || 0
      if (threshold <= 0) return
      const now = (typeof performance !== 'undefined' && performance.now) ? performance.now() : Date.now()
      if (now - this.lastLabelSwitchRAF < 16) return
      this.lastLabelSwitchRAF = now
      const shouldHide = this.zoom < threshold
      this.graph.setAutoPaint(false)
      this.graph.getNodes().forEach((n) => {
        const model = n.getModel()
        const hasLabel = !!model.originalLabel
        const currentLabel = model.label
        if (shouldHide && currentLabel) {
          this.graph.updateItem(n, { label: '' })
        } else if (!shouldHide && !currentLabel && hasLabel) {
          this.graph.updateItem(n, { label: model.originalLabel })
        }
      })
      this.graph.setAutoPaint(true)
      this.graph.paint()
    },

    // 是否大数据量
    isLargeData(data) {
      const nodesLen = (data && data.nodes && data.nodes.length) || 0
      const threshold = (this.performance && this.performance.largeThreshold) || 500
      return nodesLen >= threshold
    },

    // 按性能优化准备数据和布局
    prepareDataForPerformance(data) {
      const cloned = {
        nodes: (data.nodes || []).map(n => ({ ...n })),
        edges: (data.edges || []).map(e => ({ ...e }))
      }
      const large = this.isLargeData(cloned)
      if (large) {
        if (this.performance && this.performance.simplifyEdgeOnLarge) {
          cloned.edges.forEach(e => {
            e.style = e.style || {}
            e.style.endArrow = false
            e.style.lineWidth = 1
          })
        }
        cloned.nodes.forEach(n => {
          if (typeof n.label === 'string' && n.label) {
            n.originalLabel = n.label
            n.label = ''
          }
        })
        if (this.performance && this.performance.disableAnimateOnLarge && this.graph) {
          this.graph.updateLayout({ animate: false })
        }
      } else {
        cloned.nodes.forEach(n => {
          if (n.originalLabel && !n.label) {
            n.label = n.originalLabel
          }
        })
      }
      // 更新布局配置（含 worker / gpu）
      const layoutCfgBase = this.getLayoutConfig()
      const enableWorker = !!(this.performance && this.performance.enableWorkerLayout)
      if ((this.layout === 'force' || this.layout === 'gForce') && this.graph) {
        const extra = {
          preventOverlap: true,
          maxIteration: 500
        }
        if (this.layout === 'gForce') {
          extra.gpuEnabled = true
        }
        if (enableWorker) {
          extra.workerEnabled = true
        }
        this.graph.updateLayout({ type: this.layout, ...layoutCfgBase, ...extra })
      }
      return cloned
    },

    // 生成假数据
    generateMockData(nodeCount = 200, edgeCount = 300) {
      const nodes = []
      const edges = []
      for (let i = 0; i < nodeCount; i++) {
        nodes.push({
          id: 'n-' + i,
          label: '节点 ' + i
        })
      }
      for (let i = 0; i < edgeCount; i++) {
        const s = Math.floor(Math.random() * nodeCount)
        let t = Math.floor(Math.random() * nodeCount)
        if (t === s) t = (t + 1) % nodeCount
        edges.push({
          id: 'e-' + i,
          source: 'n-' + s,
          target: 'n-' + t,
          label: ''
        })
      }
      return { nodes, edges }
    },

    // 放大
    zoomIn() {
      if (!this.graph) return
      const currentZoom = this.graph.getZoom()
      const newZoom = Math.min(currentZoom * 1.2, 3)
      this.graph.zoomTo(newZoom)
    },

    // 缩小
    zoomOut() {
      if (!this.graph) return
      const currentZoom = this.graph.getZoom()
      const newZoom = Math.max(currentZoom * 0.8, 0.3)
      this.graph.zoomTo(newZoom)
    },

    // 适应画布
    fitView() {
      if (!this.graph) return
      // G6 v3 仅支持传 padding
      this.graph.fitView(20)
    },

    // 重置缩放
    resetZoom() {
      if (!this.graph) return
      this.graph.zoomTo(1)
      this.fitView()
    },

    // 刷新数据
    refreshData() {
      this.loading = true
      if (this.useMock) {
        const mock = this.generateMockData(this.mockNodes, this.mockEdges)
        this.updateGraphData(mock)
        this.finishRefresh()
      } else {
        this.$emit('refresh', () => {
          // 刷新完成回调
          this.loading = false
          if (this.autoFit) {
            this.$nextTick(() => {
              this.fitView()
            })
          }
        })
      }
    },
    
    // 手动完成刷新（供外部调用）
    finishRefresh() {
      this.loading = false
      if (this.autoFit && this.graph) {
        this.$nextTick(() => {
          this.fitView()
        })
      }
    },

    // 处理窗口大小变化
    handleResize() {
      if (!this.graph || !this.$refs.graphContainer) return
      
      const container = this.$refs.graphContainer
      const width = container.clientWidth
      const height = container.clientHeight
      
      this.graph.changeSize(width, height)
    },

    // 获取图谱实例（供外部调用）
    getGraph() {
      return this.graph
    },

    // 导出图片
    downloadImage(fileName = 'graph') {
      if (!this.graph) return
      
      if (typeof this.graph.downloadFullImage === 'function') {
        this.graph.downloadFullImage(fileName, 'image/png', {
          backgroundColor: '#fff',
          padding: [10, 10, 10, 10]
        })
      } else if (typeof this.graph.downloadImage === 'function') {
        this.graph.downloadImage(fileName)
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.graph-map-container {
  position: relative;
  width: calc(100% - 20px);
  height: calc(100% - 20px);
  background: #fff;
  border-radius: 4px;
  overflow: hidden;
  margin:10px;
  .graph-toolbar {
    position: absolute;
    top: 10px;
    right: 10px;
    z-index: 10;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px;
    background: rgba(255, 255, 255, 0.9);
    border: 1px solid #e4e7ed;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

    .el-button {
      margin: 0;
    }

    .el-divider {
      margin: 0 4px;
      height: 20px;
    }
  }

  .graph-container {
    width: 100%;
    height: 100%;
    min-height: 400px;
  }
}
</style>

