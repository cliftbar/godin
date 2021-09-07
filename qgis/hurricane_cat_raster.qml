<!DOCTYPE qgis PUBLIC 'http://mrcc.com/qgis.dtd' 'SYSTEM'>
<qgis version="3.20.2-Odense" minScale="1e+08" hasScaleBasedVisibilityFlag="0" maxScale="0" styleCategories="AllStyleCategories">
  <flags>
    <Identifiable>1</Identifiable>
    <Removable>1</Removable>
    <Searchable>1</Searchable>
    <Private>0</Private>
  </flags>
  <temporal enabled="0" fetchMode="0" mode="0">
    <fixedRange>
      <start></start>
      <end></end>
    </fixedRange>
  </temporal>
  <customproperties>
    <Option type="Map">
      <Option name="WMSBackgroundLayer" type="bool" value="false"/>
      <Option name="WMSPublishDataSourceUrl" type="bool" value="false"/>
      <Option name="embeddedWidgets/count" type="int" value="0"/>
      <Option name="identify/format" type="QString" value="Value"/>
    </Option>
  </customproperties>
  <pipe>
    <provider>
      <resampling enabled="false" maxOversampling="2" zoomedOutResamplingMethod="nearestNeighbour" zoomedInResamplingMethod="nearestNeighbour"/>
    </provider>
    <rasterrenderer alphaBand="-1" nodataColor="" classificationMin="0" type="singlebandpseudocolor" band="1" classificationMax="110" opacity="0.6">
      <rasterTransparency/>
      <minMaxOrigin>
        <limits>MinMax</limits>
        <extent>WholeRaster</extent>
        <statAccuracy>Estimated</statAccuracy>
        <cumulativeCutLower>0.02</cumulativeCutLower>
        <cumulativeCutUpper>0.98</cumulativeCutUpper>
        <stdDevFactor>2</stdDevFactor>
      </minMaxOrigin>
      <rastershader>
        <colorrampshader colorRampType="DISCRETE" clip="0" labelPrecision="0" classificationMode="2" minimumValue="0" maximumValue="110">
          <colorramp name="[source]" type="gradient">
            <Option type="Map">
              <Option name="color1" type="QString" value="116,182,173,0"/>
              <Option name="color2" type="QString" value="215,25,28,255"/>
              <Option name="discrete" type="QString" value="0"/>
              <Option name="rampType" type="QString" value="gradient"/>
              <Option name="stops" type="QString" value="0.309091;116,182,173,0:0.581818;183,226,168,255:0.754545;231,245,183,255:0.872727;254,232,164,255:1.02727;253,186,110,255:1.24545;237,110,67,255"/>
            </Option>
            <prop k="color1" v="116,182,173,0"/>
            <prop k="color2" v="215,25,28,255"/>
            <prop k="discrete" v="0"/>
            <prop k="rampType" v="gradient"/>
            <prop k="stops" v="0.309091;116,182,173,0:0.581818;183,226,168,255:0.754545;231,245,183,255:0.872727;254,232,164,255:1.02727;253,186,110,255:1.24545;237,110,67,255"/>
          </colorramp>
          <item alpha="0" label="&lt;= 34 kts" color="#74b6ad" value="34"/>
          <item alpha="255" label="TD (34 - 64 kts)" color="#b7e2a8" value="64"/>
          <item alpha="255" label="CAT 1 (64 - 83 kts" color="#e7f5b7" value="83"/>
          <item alpha="255" label="CAT 2 (83 - 96 kts)" color="#fee8a4" value="96"/>
          <item alpha="255" label="CAT 3 (96 - 113 kts)" color="#fdba6e" value="113"/>
          <item alpha="255" label="CAT 4 (113 - 137 kts)" color="#ed6e43" value="137"/>
          <item alpha="255" label="CAT 5 (> 137 kts)" color="#d7191c" value="inf"/>
          <rampLegendSettings orientation="2" maximumLabel="" minimumLabel="" useContinuousLegend="1" suffix="" direction="0" prefix="">
            <numericFormat id="basic">
              <Option type="Map">
                <Option name="decimal_separator" type="QChar" value=""/>
                <Option name="decimals" type="int" value="6"/>
                <Option name="rounding_type" type="int" value="0"/>
                <Option name="show_plus" type="bool" value="false"/>
                <Option name="show_thousand_separator" type="bool" value="true"/>
                <Option name="show_trailing_zeros" type="bool" value="false"/>
                <Option name="thousand_separator" type="QChar" value=""/>
              </Option>
            </numericFormat>
          </rampLegendSettings>
        </colorrampshader>
      </rastershader>
    </rasterrenderer>
    <brightnesscontrast brightness="0" gamma="1" contrast="0"/>
    <huesaturation colorizeRed="255" grayscaleMode="0" saturation="0" colorizeStrength="100" colorizeGreen="128" colorizeBlue="128" colorizeOn="0"/>
    <rasterresampler maxOversampling="2"/>
    <resamplingStage>resamplingFilter</resamplingStage>
  </pipe>
  <blendMode>0</blendMode>
</qgis>
