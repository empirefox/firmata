import 'dart:async';
import 'dart:developer';

import 'package:flutter_easyloading/flutter_easyloading.dart';
import 'package:get/get.dart' show Get, Inst;
import 'package:grpc/grpc.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/data/services/storage.service.dart';
import 'package:monolith/app/share/ext.dart';
import 'package:monolith/objectbox.g.dart';
import 'package:monolith/pb/empirefox/firmata/board.pb.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';
import 'package:monolith/pb/empirefox/firmata/instance.pb.dart';
import 'package:monolith/pb/empirefox/firmata/integration.pb.dart';
import 'package:monolith/pb/empirefox/firmata/mode.pb.dart';
import 'package:monolith/pb/empirefox/firmata/pinname.pb.dart';
import 'package:monolith/pb/empirefox/firmata/transport.pbgrpc.dart';
import 'package:monolith/pb/google/protobuf/empty.pb.dart';
import 'package:protobuf/protobuf.dart';

const _options = ChannelOptions();

typedef TypedAsyncCallback<T> = Future<void> Function(T t);

class TypedServerMessage {
  final ServerMessage_Type type;
  final Instance? connected;
  final int? disconnected;
  final InstancePins? digital;
  final AnalogMessage? analog;
  const TypedServerMessage(
      {required this.type,
      this.connected,
      this.disconnected,
      this.digital,
      this.analog});

  static const _invalid = TypedServerMessage(type: ServerMessage_Type.notSet);
  bool get isValid => type != ServerMessage_Type.notSet;
  bool get isInvalid => type == ServerMessage_Type.notSet;
}

class InstancePins {
  final int firmata;
  final List<int> pins;
  const InstancePins(this.firmata, this.pins);
  static const _invalid = InstancePins(-1, []);
  bool get isValid => this != _invalid;
  bool get isInvalid => this == _invalid;
}

class AnalogMessage {
  final int firmata;
  final int pin;
  final int value;
  const AnalogMessage(this.firmata, this.pin, this.value);
  static const _invalid = AnalogMessage(-1, 0, 0);
  bool get isValid => this != _invalid;
  bool get isInvalid => this == _invalid;
}

class _Pair<T> {
  final T o;
  final T t;
  _Pair(this.o, this.t);
  static List<T> createPairs<T>() => <T>[];
}

class Transport {
  final PlanetConfig planetConfig;
  late final ClientChannel _channel;
  late final TransportClient _client;

  final TypedAsyncCallback<Transport>? onAboutToClose;
  final TypedAsyncCallback<Transport>? onClosed;

  late final Version_Peer apiVersion;
  late final List<Board> boards;
  late final Map<String, Board> boardById;

  late final Integration integration;
  late final Integration originalIntegration;
  final Map<int, List<_Pair<Wiring_FirmataPins>>> _wirePinsPairsByFirmata = {};

  late final Config config;
  late final Config originalConfig;
  final List<List<Group_Pin>> groupVisiblePins = [];
  final Map<int, List<_Pair<Group_Pin>>> _groupPinPairsByFirmata = {};
  final Map<int, List<_Pair<Group_DigitalInputPin>>> _detectPinPairsByFirmata =
      {};

  late final List<Instance?> instances;
  late final Stream<TypedServerMessage> onServerMessage;

  Stream<Instance> get onConnected => onServerMessage
      .takeWhile((msg) => msg.type == ServerMessage_Type.connected)
      .map((msg) => msg.connected!);

  Stream<int> get onDisconnected => onServerMessage
      .takeWhile((msg) => msg.type == ServerMessage_Type.disconnected)
      .map((msg) => msg.disconnected!);

  Stream<InstancePins> get onDigitalMessage => onServerMessage
      .takeWhile((msg) => msg.type == ServerMessage_Type.digital)
      .map((msg) => msg.digital!);

  Stream<AnalogMessage> get onAnalogMessage => onServerMessage
      .takeWhile((msg) => msg.type == ServerMessage_Type.analog)
      .map((msg) => msg.analog!);

  static Future<Transport> create(
    PlanetConfig planetConfig, {
    Iterable<ClientInterceptor>? interceptors,
    TypedAsyncCallback<Transport>? onAboutToClose,
    TypedAsyncCallback<Transport>? onClosed,
  }) async {
    final t = Transport._(planetConfig,
        interceptors: interceptors,
        onAboutToClose: onAboutToClose,
        onClosed: onClosed);
    await t._init();
    return t;
  }

  Transport._(
    this.planetConfig, {
    Iterable<ClientInterceptor>? interceptors,
    this.onAboutToClose,
    this.onClosed,
  }) {
    _channel = ClientChannel(
      planetConfig.host,
      port: planetConfig.port,
      options: ChannelOptions(
        credentials: planetConfig.isTlsDisabled
            ? const ChannelCredentials.insecure()
            : ChannelCredentials.secure(
                certificates:
                    planetConfig.tlsCertificates.zeroAsNull?.codeUnits,
                password: planetConfig.tlsPassword.zeroAsNull,
                authority: planetConfig.tlsAuthority.zeroAsNull,
                onBadCertificate: planetConfig.canTlsInsecureSkipVerify
                    ? allowBadCertificates
                    : null),
        codecRegistry: CodecRegistry(codecs: [
          if (planetConfig.supportGrpcCodecGzip) const GzipCodec(),
          if (planetConfig.supportGrpcCodecIdentity) const IdentityCodec()
        ]),
        connectionTimeout: Duration(
          seconds: planetConfig.connectionTimeoutSeconds,
        ),
        idleTimeout: Duration(
          seconds: planetConfig.idleTimeoutSeconds,
        ),
        userAgent: planetConfig.userAgent.zeroAsNull ?? _options.userAgent,
      ),
    );

    _client = TransportClient(
      _channel,
      options: CallOptions(
        timeout: Duration(
          seconds: planetConfig.callTimeoutSeconds,
        ),
        providers: [
          // TODO with github.com/grpc-ecosystem/go-grpc-middleware/auth?
          authenticate(planetConfig),
        ],
      ),
      interceptors: interceptors,
    );
  }

  static MetadataProvider authenticate(PlanetConfig config) {
    if (config.tokenType == 'none') {
      return (Map<String, String> metadata, String uri) async {};
    }
    final boxes = Get.find<StorageService>();
    final q = boxes.planet.query(PlanetConfig_.id.equals(config.id)).build()
      ..limit = 1;
    return (Map<String, String> metadata, String uri) async {
      final r = q.property(PlanetConfig_.token).find();
      config.token = r.isNotEmpty ? r.first : '';
      if (config.token.isNotEmpty) {
        metadata['authorization'] = '${config.tokenType} ${config.token}';
      }
    };
  }

  Future<void> shutdown() async {
    preCall();
    await onAboutToClose?.call(this);
    await _channel.shutdown();
    await onClosed?.call(this);
    postCall();
  }

  Future<void> terminate() async {
    preCall();
    await onAboutToClose?.call(this);
    await _channel.terminate();
    await onClosed?.call(this);
    postCall();
  }

  Future<void> _init() async {
    apiVersion = await _client.getApiVersion(Empty.getDefault());
    boards = (await _client.listBoards(Empty.getDefault())).boards;
    boardById = Map.fromIterable(boards, key: (b) => (b as Board).id);
    await _initFromIntegeration();
    instances = List.filled(integration.firmatas.length, null);
    await _initFromConfig();

    onServerMessage = _onServerMessage().asBroadcastStream();
    onServerMessage.listen(null, cancelOnError: true);
  }

  Future<void> _initFromIntegeration() async {
    originalIntegration = await _client.getIntegration(Empty.getDefault());
    integration = originalIntegration.deepCopy();
    final oFirmatas = originalIntegration.firmatas;
    final tFirmatas = integration.firmatas;
    final totalFirmatas = oFirmatas.length;
    for (var i = 0; i < totalFirmatas; i++) {
      final oWires = oFirmatas[i].wiring;
      final tWires = tFirmatas[i].wiring;
      final totalWires = oWires.length;
      for (var j = 0; j < totalWires; j++) {
        final oWire = oWires[j];
        final tWire = tWires[j];
        final oFrom = oWire.from;
        final tFrom = tWire.from;
        final oTo = oWire.to;
        final tTo = tWire.to;
        _wirePinsPairsByFirmata
            .putIfAbsent(oFrom.firmataIndex, _Pair.createPairs)
            .add(_Pair(oFrom, tFrom));
        if (oTo.hasFirmata()) {
          _wirePinsPairsByFirmata
              .putIfAbsent(oTo.firmata.firmataIndex, _Pair.createPairs)
              .add(_Pair(oTo.firmata, tTo.firmata));
        }
      }
    }

    final oDevices = originalIntegration.devices;
    final tDevices = integration.devices;
    final totalDevices = originalIntegration.devices.length;
    for (var i = 0; i < totalDevices; i++) {
      final oWires = oDevices[i].wiring;
      final tWires = tDevices[i].wiring;
      final totalWires = oWires.length;
      for (var j = 0; j < totalWires; j++) {
        final oWire = oWires[j];
        final tWire = tWires[j];
        final oTo = oWire.to;
        final tTo = tWire.to;
        if (oTo.hasFirmata()) {
          _wirePinsPairsByFirmata
              .putIfAbsent(oTo.firmata.firmataIndex, _Pair.createPairs)
              .add(_Pair(oTo.firmata, tTo.firmata));
        }
      }
    }
  }

  Future<void> _initFromConfig() async {
    originalConfig = await _client.getConfig(Empty.getDefault());
    config = originalConfig.deepCopy();
    final oGroups = originalConfig.groups;
    final tGroups = config.groups;
    final totalGroups = oGroups.length;
    for (var i = 0; i < totalGroups; i++) {
      final oPins = oGroups[i].pins;
      final tPins = tGroups[i].pins;
      final totalPins = oPins.length;
      int offset = 0;
      for (var j = 0; j < totalPins; j++) {
        final idx = j + offset;
        final oPin = oPins[idx];
        final tPin = tPins[idx];
        _groupPinPairsByFirmata
            .putIfAbsent(oPin.firmataIndex, _Pair.createPairs)
            .add(_Pair(oPin, tPin));

        if (oPin.switch_21.hasDetect()) {
          final o = oPin.switch_21.detect;
          final t = tPin.switch_21.detect;
          _detectPinPairsByFirmata
              .putIfAbsent(o.firmataIndex, _Pair.createPairs)
              .add(_Pair(o, t));
        }
      }
      groupVisiblePins[i] = tGroups[i].visiblePins;
    }
  }

  Stream<TypedServerMessage> _onServerMessage() {
    return _client.onServerMessage(Empty.getDefault()).map((msg) {
      switch (msg.whichType()) {
        case ServerMessage_Type.connected:
          final inst = _onConnected(msg.connected);
          instances[inst.firmataIndex] = inst;
          return TypedServerMessage(
            type: ServerMessage_Type.connected,
            connected: inst,
          );
        case ServerMessage_Type.disconnected:
          final firmataIndex = _onDisConnected(msg.disconnected);
          instances[firmataIndex] = null;
          return TypedServerMessage(
            type: ServerMessage_Type.disconnected,
            disconnected: firmataIndex,
          );
        case ServerMessage_Type.digital:
          final instancePins = _onDigitalMessage(msg.digital);
          if (instancePins.isInvalid) return TypedServerMessage._invalid;
          return TypedServerMessage(
            type: ServerMessage_Type.digital,
            digital: instancePins,
          );
        case ServerMessage_Type.analog:
          final analogMessage = _onAnalogMessage(msg.analog);
          if (analogMessage.isInvalid) return TypedServerMessage._invalid;
          return TypedServerMessage(
            type: ServerMessage_Type.analog,
            analog: analogMessage,
          );
        case ServerMessage_Type.notSet:
          log('ServerMessage.type required: ${msg.toString()}');
          return TypedServerMessage._invalid;
      }
    }).takeWhile((msg) => msg.isValid);
  }

  Instance _onConnected(Instance inst) {
    final dxByName = <PinName, int>{};
    final dxByAx = <int, int>{};
    for (var pin in inst.pins) {
      final dx = pin.dx;
      dxByName[pin.name] = dx;
      if (pin.ax != 127) dxByAx[pin.ax] = dx;
    }
    _wirePinsPairsByFirmata.remove(inst.firmata)?.forEach((pair) {
      switch (pair.o.whichFirst()) {
        case Wiring_FirmataPins_First.gpioName:
          pair.t.dx = dxByName[pair.o.gpioName] ?? -1;
          break;
        case Wiring_FirmataPins_First.dx:
          break;
        case Wiring_FirmataPins_First.ax:
          pair.t.dx = dxByAx[pair.o.ax] ?? -1;
          break;
        case Wiring_FirmataPins_First.notSet:
          break;
      }

      switch (pair.o.whichSlice()) {
        case Wiring_FirmataPins_Slice.lastGpioName:
          pair.t.lastDx = dxByName[pair.o.gpioName] ?? -1;
          break;
        case Wiring_FirmataPins_Slice.lastDx:
          break;
        case Wiring_FirmataPins_Slice.lastAx:
          pair.t.lastDx = dxByAx[pair.o.ax] ?? -1;
          break;
        case Wiring_FirmataPins_Slice.notSet:
          break;
      }
    });

    _groupPinPairsByFirmata.remove(inst.firmata)?.forEach((pair) {
      switch (pair.o.whichId()) {
        case Group_Pin_Id.gpioName:
          pair.t.dx = dxByName[pair.o.gpioName] ?? -1;
          break;
        case Group_Pin_Id.dx:
          break;
        case Group_Pin_Id.ax:
          pair.t.dx = dxByAx[pair.o.ax] ?? -1;
          break;
        case Group_Pin_Id.notSet:
          break;
      }
    });

    _detectPinPairsByFirmata.remove(inst.firmata)?.forEach((pair) {
      switch (pair.o.whichId()) {
        case Group_DigitalInputPin_Id.gpioName:
          pair.t.dx = dxByName[pair.o.gpioName] ?? -1;
          break;
        case Group_DigitalInputPin_Id.dx:
          break;
        case Group_DigitalInputPin_Id.ax:
          pair.t.dx = dxByAx[pair.o.ax] ?? -1;
          break;
        case Group_DigitalInputPin_Id.notSet:
          break;
      }
    });
    return inst;
  }

  int _onDisConnected(int firmataIndex) {
    return firmataIndex;
  }

  InstancePins _onDigitalMessage(ServerMessage_Digital r) {
    final firmataIndex = r.firmata;
    final inst = instances[firmataIndex];
    if (inst == null) return InstancePins._invalid;

    final rport = r.port & 0xFF;
    final rpins = r.pins & 0xFF;
    final rvalues = r.values & 0xFF;
    if (rpins == 0) return InstancePins._invalid;

    final totalPins = inst.pins.length;
    final pins = <int>[];
    for (var i = 0; i < 8; i++) {
      final bit = (1 << i);
      final dx = rport * 8 + i;
      if (dx >= totalPins) break;
      if ((rpins & bit) == 0) continue;
      final value = ((rvalues & bit) == 0) ? 0 : 1;
      inst.pins[dx].value = value;
      pins.add(dx);
    }

    if (pins.isEmpty) return InstancePins._invalid;
    return InstancePins(firmataIndex, pins);
  }

  AnalogMessage _onAnalogMessage(ServerMessage_Analog r) {
    final msg = AnalogMessage(r.firmata, r.pin, r.value);
    final inst = instances[msg.firmata];
    if (inst == null || msg.pin >= inst.pins.length) {
      return AnalogMessage._invalid;
    }
    inst.pins[msg.pin].value = msg.value;
    return msg;
  }

  Future<void> connect(int firmataIndex) {
    preCall();
    return _client
        .connect(FirmataIndex.create()..firmata = firmataIndex)
        .whenComplete(postCall);
  }

  Future<void> disconnect(int firmataIndex) {
    preCall();
    return _client
        .connect(FirmataIndex.create()..firmata = firmataIndex)
        .whenComplete(postCall);
  }

  Future<void> preCall() => EasyLoading.show(status: 'sending...');
  Future<void> postCall() => EasyLoading.dismiss();

  Future<void> setPinMode(
      {required int firmata, required int dx, required Mode mode}) {
    preCall();
    return _client
        .setPinMode(SetPinModeRequest.create()
          ..firmata = firmata
          ..dx = dx
          ..mode = mode)
        .whenComplete(postCall);
  }

  Future<void> triggerDigitalPin(
      {required int group, required int gpin, int? realtimeTriggerMs}) {
    preCall();
    return _client
        .triggerDigitalPin(TriggerDigitalPinRequest.create()
          ..group = group
          ..gpin = gpin
          ..realtimeTriggerMs = realtimeTriggerMs ?? 0)
        .whenComplete(postCall);
  }

  Future<void> setPinValue(
      {required int group, required int gpin, required int value}) {
    preCall();
    return _client
        .setPinValue(SetPinValueRequest.create()
          ..group = group
          ..gpin = gpin
          ..value = value)
        .whenComplete(() => EasyLoading.dismiss());
  }

  Future<void> reportDigital(
      {required int firmata, required int port, required bool enable}) {
    preCall();
    return _client
        .reportDigital(ReportDigitalRequest.create()
          ..firmata = firmata
          ..port = port
          ..enable = enable)
        .whenComplete(() => EasyLoading.dismiss());
  }

  Future<void> reportAnalog(
      {required int firmata, required int pin, required bool enable}) {
    preCall();
    return _client
        .reportAnalog(ReportAnalogRequest.create()
          ..firmata = firmata
          ..pin = pin
          ..enable = enable)
        .whenComplete(() => EasyLoading.dismiss());
  }

  Future<void> writeString({required int firmata, required String data}) {
    preCall();
    return _client
        .writeString(WriteStringRequest.create()
          ..firmata = firmata
          ..data = data)
        .whenComplete(() => EasyLoading.dismiss());
  }

  Future<void> setSamplingInterval({required int firmata, required int ms}) {
    preCall();
    return _client
        .setSamplingInterval(SetSamplingIntervalRequest.create()
          ..firmata = firmata
          ..ms = ms)
        .whenComplete(() => EasyLoading.dismiss());
  }
}

extension TransportGroupExt on Group {
  bool isNotVisible(Group_Pin pin) {
    final type = pin.whichType();
    return type == Group_Pin_Type.notSet ||
        (type == Group_Pin_Type.hide && pin.hide);
  }

  List<Group_Pin> get visiblePins => pins.skipWhile(isNotVisible).toList();
}

extension TransportGroupPinExt on Group_Pin {
  Instance_Pin? getPin(Transport t) => t.instances[firmataIndex]?.pins[dx];

  Instance_Pin? getDetectPin(Transport t) {
    if (whichType() != Group_Pin_Type.switch_21) return null;
    final type = switch_21;
    if (!type.hasDetect()) return null;
    final detect = switch_21.detect;
    return t.instances[detect.firmataIndex]?.pins[detect.dx];
  }
}

extension TransportInstancePinExt on Instance_Pin {
  bool get isDigital => mode != Mode.SERIAL;
  bool get isSerial => mode == Mode.SERIAL;
  bool get isAnalog => ax != 127;
}