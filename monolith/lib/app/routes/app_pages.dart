import 'package:get/get.dart';

import 'package:monolith/app/modules/about/bindings/about_binding.dart';
import 'package:monolith/app/modules/about/views/about_view.dart';
import 'package:monolith/app/modules/home/bindings/home_binding.dart';
import 'package:monolith/app/modules/home/views/home_view.dart';
import 'package:monolith/app/modules/planet/bindings/planet_binding.dart';
import 'package:monolith/app/modules/planet/views/planet_view.dart';
import 'package:monolith/app/modules/planet_edit/bindings/planet_edit_binding.dart';
import 'package:monolith/app/modules/planet_edit/views/planet_edit_view.dart';
import 'package:monolith/app/modules/planets/bindings/planets_binding.dart';
import 'package:monolith/app/modules/planets/views/planets_view.dart';

part 'app_routes.dart';

class AppPages {
  AppPages._();

  static const INITIAL = Routes.PLANETS;

  static final routes = [
    GetPage(
      name: _Paths.HOME,
      page: () => HomeView(),
      binding: HomeBinding(),
    ),
    GetPage(
      name: _Paths.PLANET,
      page: () => PlanetView(),
      binding: PlanetBinding(),
    ),
    GetPage(
      name: _Paths.PLANETS,
      page: () => PlanetsView(),
      binding: PlanetsBinding(),
    ),
    GetPage(
      name: _Paths.PLANET_EDIT,
      page: () => PlanetEditView(),
      binding: PlanetEditBinding(),
    ),
    GetPage(
      name: _Paths.ABOUT,
      page: () => AboutView(),
      binding: AboutBinding(),
    ),
  ];
}
