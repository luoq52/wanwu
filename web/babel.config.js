module.exports = {
  presets: [
    ['@babel/preset-env', { 
        useBuiltIns: 'usage', 
        corejs: 3, 
        shippedProposals: true,
        useBuiltIns:false
     }]
  ],
  plugins: [
    ['@babel/plugin-proposal-class-properties', { loose: true }],
    ['@babel/plugin-proposal-private-methods', { loose: true }],
    ['@babel/plugin-proposal-private-property-in-object', { loose: true }]
  ]
}

