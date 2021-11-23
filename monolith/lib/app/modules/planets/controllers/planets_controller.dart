import 'package:get/get.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/data/services/storage.service.dart';
import 'package:monolith/app/data/services/transport.service.dart';
import 'package:monolith/app/routes/app_pages.dart';

class PlanetsController extends GetxController {
  final _ss = Get.find<StorageService>();
  final _ts = Get.find<TransportService>();

  late List<PlanetConfig> planets;

  @override
  void onInit() {
    super.onInit();
    planets = _ss.planet.getAll();
  }

  @override
  void onReady() {
    super.onReady();
  }

  @override
  void onClose() {}

  bool isOnline(PlanetConfig planet) => _ts.isOnline(planet.id);

  Future<String>? findServerName(int id) =>
      _ts.getOrNull(id)?.then((t) => t.config.nick);

  void about() => Get.toNamed(Routes.ABOUT);

  Future<void> create() async {
    final r = await Get.toNamed(
      '${Routes.PLANET_EDIT}?id=0',
      preventDuplicates: false,
    );
    if (r == null) return;
    final PlanetConfig n = r;
    planets.insert(0, n);
    update(['planets']);
  }

  void delete(PlanetConfig planet) {
    _ss.planet.remove(planet.id);
    planets.remove(planet);
    update(['planets']);
  }

  Future<void> edit(PlanetConfig planet) async {
    final r = await Get.toNamed(
      '${Routes.PLANET_EDIT}?id=${planet.id}',
      arguments: planet,
    );
    if (r == null) return;
    final PlanetConfig n = r;
    planets[planets.indexOf(planet)] = n;
    update([n.id]);
  }

  Future<void> shutdown(PlanetConfig planet) async {
    await _ts.shutdown(planet.id);
    update([planet.id]);
  }

  Future<void> view(PlanetConfig planet) async {
    final r = await Get.toNamed(
      '${Routes.PLANET}?id=${planet.id}',
      arguments: planet,
    );
    final bool? isDown = r;
    if (isDown == true) {
      await _ts.shutdown(planet.id);
      update([planet.id]);
    }
  }
}
