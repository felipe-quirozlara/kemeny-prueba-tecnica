## Problemas críticos

### Iniciar proyecto
Problema: El levantamiento del proyecto está incorrecto, porque el archivo init.sql contiene error con columnas UUID. Crea un tabla con UUID y luego al insertar los datos en la base de datos, los UUID no son validos, lo que evita poder levantar el proyecto. Esto es un problema criticos para el equipo pueda revisar el proyecto

Solución: Modificar el init.sql para la creación automatica de UUID por parte de Postgresql para la tabla "Tags".

### JWT invalido
Problema: El proyecto genera JWT token invalidos en el campo EXP usando fmt.Sprintf para crear la fecha de expiración del token, pero esto es invalido para luego validar el token

Solución: Eliminar la función fmt.Sprintf para dejar el valor de expiración como unix timestamp en lugar del string que genera el método fmt.Sprintf


### Front end invalido
Problema: El front utiliza multiples endpoint desde el servicio REST, pero la página no tiene ninguna forma para que el usuario inicie sesión, por consecuencia el usuario siempré verá las páginas con errores para obtener la información a menos que modifique sus archivos en localStorage para poner un JWT valido.

Solución: Agregar una página para el inicio de sesión (sin querer salir del scope de la prueba) para un proyecto es critico poder revisar el flujo completo y la falta de este login solo permitiría a los usuario que hayan detectado el problema.

### Rendimiento endpoint ListTasks
problema: El endpoint ListTasks tiene un doble bucle que primero obtiene todas las tareas, después obtiene a quien están asignadas y luego otro bucle para las etiquetas. Esto es un problema de rendimiento porque cada consulta realiza dos trabajos.

Solución: Optimizar consulta de sql utilizando JOIN para obtener toda la información necesaria en una consulta, por ejemplo utilizando JOINS.

### Falta control de autorización en handlers de tareas
Problema: Los handlers de tareas permiten leer y modificar recursos sin verificar correctamente la propiedad o el rol del usuario. Actualmente se valida el JWT para autenticación, pero no se compara el `user_id` del token con el propietario/assignado de la tarea ni se aplican comprobaciones de rol (admin vs member). Esto permite que un usuario autenticado pueda consultar o actualizar tareas que no le pertenecen.

Solución: Añadir comprobaciones de autorización en los endpoints de lectura/actualización/eliminación de tareas — comparar el `user_id` extraído del JWT con el `owner_id` o `assignee_id` de la tarea y denegar el acceso con 403 cuando corresponda; introducir roles (admin/member) y validar permisos para operaciones administrativas. Agregar tests que comprueben acceso autorizado y denegado.

### Centralizar errores en `tasks.go` y evitar enviar texto plano
Problema: Los handlers en `tasks.go` retornan errores como texto plano o mensajes internos (por ejemplo `http.Error(w, err.Error(), ...)` o `w.Write([]byte("..."))`) y cada handler replica la misma lógica de respuesta. Esto provoca respuestas inconsistentes (diferentes formatos y códigos HTTP), exposición de mensajes internos al cliente y dificultad para instrumentar y testear el manejo de errores.

Solución: Implementar un middleware o helper centralizado para respuestas de error que devuelva JSON estandarizado (por ejemplo `{ "error": "mensaje", "code": 123 }`) y mapee errores internos a códigos HTTP apropiados. Evitar exponer `err.Error()` directamente al cliente; registrar el error completo en logs y devolver mensajes amigables.
