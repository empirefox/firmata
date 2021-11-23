import 'package:dartx/dartx.dart' show StringToIntOrNullExtension;
import 'package:get/get.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/data/services/storage.service.dart';
import 'package:monolith/app/data/services/transport.service.dart';
import 'package:monolith/app/routes/app_pages.dart';
import 'package:monolith/app/share/ext.dart';
import 'package:reactive_forms/reactive_forms.dart';

class PlanetEditController extends GetxController {
  final _ss = Get.find<StorageService>();
  final _ts = Get.find<TransportService>();
  final _default = PlanetConfig.defaultValue;

  late int id;
  late PlanetConfig config;
  late FormGroup form;

  late bool invalid = false;

  @override
  void onInit() {
    super.onInit();
    _initConfig();
  }

  void _initConfig() {
    final id = Get.parameters['id']?.toIntOrNull();
    if (id == null) {
      invalid = true;
      return;
    }

    this.id = id;
    var config = _ss.getxPlanetOrFromId(id);
    if (config == null && id == 0) config = PlanetConfig();
    if (config == null) {
      invalid = true;
      return;
    }
    this.config = config;

    if (!invalid) {
      _initForm();
    }
  }

  void _initForm() {
    form = FormGroup({
      'id': FormControl<int>(
        value: config.id,
      ),
      'name': FormControl<String>(
        value: config.name.zeroAsNull ?? _default.name,
        validators: [Validators.maxLength(24)],
      ),
      'host': FormControl<String>(
        value: config.host.zeroAsNull ?? _default.host,
        validators: [Validators.required],
      ),
      'port': FormControl<int>(
        value: config.port.zeroAsNull ?? _default.port,
        validators: [
          Validators.min(1),
          Validators.max(1 << 16 - 1),
        ],
      ),
      'userAgent': FormControl<String>(
        value: config.userAgent.zeroAsNull ?? _default.userAgent,
      ),
      'isTlsDisabled': FormControl<bool>(value: config.isTlsDisabled),
      'tlsCertificates': FormControl<String>(
        value: config.tlsCertificates.zeroAsNull ?? _default.tlsCertificates,
      ),
      'tlsPassword': FormControl<String>(
        value: config.tlsPassword.zeroAsNull ?? _default.tlsPassword,
      ),
      'tlsAuthority': FormControl<String>(
        value: config.tlsAuthority.zeroAsNull ?? _default.tlsAuthority,
      ),
      'canTlsInsecureSkipVerify':
          FormControl<bool>(value: config.canTlsInsecureSkipVerify),
      'supportGrpcCodecGzip':
          FormControl<bool>(value: config.supportGrpcCodecGzip),
      'supportGrpcCodecIdentity':
          FormControl<bool>(value: config.supportGrpcCodecIdentity),
      'tokenType': FormControl<String>(
        value: config.tokenType.zeroAsNull ?? _default.tokenType,
      ),
      'token': FormControl<String>(
        value: config.token.zeroAsNull ?? _default.token,
      ),
      'connectionTimeoutSeconds': FormControl<int>(
        value: config.connectionTimeoutSeconds.zeroAsNull ??
            _default.connectionTimeoutSeconds,
        validators: [Validators.min(1)],
      ),
      'idleTimeoutSeconds': FormControl<int>(
        value:
            config.idleTimeoutSeconds.zeroAsNull ?? _default.idleTimeoutSeconds,
        validators: [Validators.min(1)],
      ),
      'callTimeoutSeconds': FormControl<int>(
        value:
            config.callTimeoutSeconds.zeroAsNull ?? _default.callTimeoutSeconds,
        validators: [Validators.min(1)],
      ),
    });
  }

  @override
  void onReady() {
    super.onReady();
  }

  @override
  void onClose() {}

  void about() => Get.toNamed(Routes.ABOUT);

  void onSubmit() async {
    final n = PlanetConfig.fromJson(form.value);
    final id = await _ss.planet.putAsync(n);
    await _ts.shutdown(id);
    Get.back(result: n, closeOverlays: true);
  }
}
