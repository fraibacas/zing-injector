// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: metric/metric.proto

package org.zenoss.zing.proto.metric;

public interface MetricOrBuilder extends
    // @@protoc_insertion_point(interface_extends:metric.Metric)
    com.google.protobuf.MessageOrBuilder {

  /**
   * <pre>
   * The metric name
   * </pre>
   *
   * <code>string metric = 1;</code>
   */
  java.lang.String getMetric();
  /**
   * <pre>
   * The metric name
   * </pre>
   *
   * <code>string metric = 1;</code>
   */
  com.google.protobuf.ByteString
      getMetricBytes();

  /**
   * <pre>
   * The time at which the value was captured
   * </pre>
   *
   * <code>int64 timestamp = 2;</code>
   */
  long getTimestamp();

  /**
   * <pre>
   * The metric value
   * </pre>
   *
   * <code>double value = 3;</code>
   */
  double getValue();

  /**
   * <pre>
   * Identifies the MetricInstance that owns this datapoint
   * </pre>
   *
   * <code>string id = 4;</code>
   */
  java.lang.String getId();
  /**
   * <pre>
   * Identifies the MetricInstance that owns this datapoint
   * </pre>
   *
   * <code>string id = 4;</code>
   */
  com.google.protobuf.ByteString
      getIdBytes();

  /**
   * <pre>
   * Dimensions associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; dimensions = 5;</code>
   */
  int getDimensionsCount();
  /**
   * <pre>
   * Dimensions associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; dimensions = 5;</code>
   */
  boolean containsDimensions(
      java.lang.String key);
  /**
   * Use {@link #getDimensionsMap()} instead.
   */
  @java.lang.Deprecated
  java.util.Map<java.lang.String, java.lang.String>
  getDimensions();
  /**
   * <pre>
   * Dimensions associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; dimensions = 5;</code>
   */
  java.util.Map<java.lang.String, java.lang.String>
  getDimensionsMap();
  /**
   * <pre>
   * Dimensions associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; dimensions = 5;</code>
   */

  java.lang.String getDimensionsOrDefault(
      java.lang.String key,
      java.lang.String defaultValue);
  /**
   * <pre>
   * Dimensions associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; dimensions = 5;</code>
   */

  java.lang.String getDimensionsOrThrow(
      java.lang.String key);

  /**
   * <pre>
   * Metadata associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; metadata = 6;</code>
   */
  int getMetadataCount();
  /**
   * <pre>
   * Metadata associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; metadata = 6;</code>
   */
  boolean containsMetadata(
      java.lang.String key);
  /**
   * Use {@link #getMetadataMap()} instead.
   */
  @java.lang.Deprecated
  java.util.Map<java.lang.String, java.lang.String>
  getMetadata();
  /**
   * <pre>
   * Metadata associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; metadata = 6;</code>
   */
  java.util.Map<java.lang.String, java.lang.String>
  getMetadataMap();
  /**
   * <pre>
   * Metadata associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; metadata = 6;</code>
   */

  java.lang.String getMetadataOrDefault(
      java.lang.String key,
      java.lang.String defaultValue);
  /**
   * <pre>
   * Metadata associated with this datapoint.
   * </pre>
   *
   * <code>map&lt;string, string&gt; metadata = 6;</code>
   */

  java.lang.String getMetadataOrThrow(
      java.lang.String key);
}
