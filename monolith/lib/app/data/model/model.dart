import 'dart:io';

import 'package:json_annotation/json_annotation.dart';
import 'package:monolith/app/share/ext.dart';
import 'package:monolith/app/share/utils.dart';
import 'package:objectbox/objectbox.dart'
    show Entity, Index, IndexType, Transient;

part 'model.g.dart';

/// flutter pub run build_runner build --delete-conflicting-outputs
@Entity()
@JsonSerializable()
class PlanetConfig implements ArgToParamer {
  @Transient()
  static final defaultValue = PlanetConfig();

  PlanetConfig();

  factory PlanetConfig.fromJson(Map<String, dynamic> json) =>
      _$PlanetConfigFromJson(json);

  /// Object IDs can not be:
  /// - 0 (zero) or null (if using java.lang.Long) As said above, when putting an
  /// object with ID zero it will be assigned an unused ID (not zero).
  /// - 0xFFFFFFFFFFFFFFFF (-1 in Java) Reserved for internal use.
  ///
  /// For example, if there is an object with ID 1 and another with ID 100 in a
  /// box, the next new object that is put will be assigned ID 101.
  int id = 0;

  @Index(type: IndexType.value)
  String name = '';

  @Index(type: IndexType.value)
  String host = '127.0.0.1';

  int port = 2525;

  bool isTlsDisabled = false;

  /// With server authentication SSL/TLS, trustedRoot: roots.pem
  String tlsCertificates = '';
  String tlsPassword = '';
  String tlsAuthority = '';
  bool canTlsInsecureSkipVerify = false;

  bool supportGrpcCodecGzip = true;
  bool supportGrpcCodecIdentity = true;

  /// ignore if set to none
  String tokenType = 'Bearer';
  String token = '';

  int connectionTimeoutSeconds = 50 * 60;
  int idleTimeoutSeconds = 5 * 60;
  int callTimeoutSeconds = 30;

  String userAgent = '';

  String viewName([String? backport]) =>
      name.zeroAsNull ?? backport?.zeroAsNull ?? address;

  @Transient()
  String get address =>
      InternetAddress.tryParse(host)?.type == InternetAddressType.IPv6
          ? '[$host]:$port'
          : '$host:$port';

  @override
  String argToParam() => id.toString();
}
