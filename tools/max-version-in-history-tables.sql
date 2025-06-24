-- SELECT 'FS_ContainerRelate' AS tablename, MAX(version) FROM History_ContainerRelate;
-- SELECT 'FS_FieldScoutingLog' AS tablename, MAX(version) FROM History_FieldScoutingLog;
-- SELECT 'FS_HabitatRelate' AS tablename, MAX(version) FROM History_HabitatRelate;
-- SELECT 'FS_InspectionSample' AS tablename, MAX(version) FROM History_InspectionSample;
-- SELECT 'FS_InspectionSampleDetail' AS tablename, MAX(version) FROM History_InspectionSampleDetail;
-- SELECT 'FS_LineLocation' AS tablename, MAX(version) FROM History_LineLocation;
-- SELECT 'FS_LocationTracking' AS tablename, MAX(version) FROM History_LocationTracking;
-- SELECT 'FS_MosquitoInspection' AS tablename, MAX(version) FROM History_MosquitoInspection;
-- SELECT 'FS_PointLocation' AS tablename, MAX(version) FROM History_PointLocation;
-- SELECT 'FS_PolygonLocation' AS tablename, MAX(version) FROM History_PolygonLocation;
-- SELECT 'FS_Pool' AS tablename, MAX(version) FROM History_Pool;
-- SELECT 'FS_PoolDetail' AS tablename, MAX(version) FROM History_PoolDetail;
-- SELECT 'FS_ProposedTreatmentArea' AS tablename, MAX(version) FROM History_ProposedTreatmentArea;
-- SELECT 'FS_QAMosquitoInspection' AS tablename, MAX(version) FROM History_QAMosquitoInspection;
-- SELECT 'FS_RodentLocation' AS tablename, MAX(version) FROM History_RodentLocation;
-- SELECT 'FS_SampleCollection' AS tablename, MAX(version) FROM History_SampleCollection;
-- SELECT 'FS_SampleLocation' AS tablename, MAX(version) FROM History_SampleLocation;
-- SELECT 'FS_ServiceRequest' AS tablename, MAX(version) FROM History_ServiceRequest;
-- SELECT 'FS_SpeciesAbundance' AS tablename, MAX(version) FROM History_SpeciesAbundance;
-- SELECT 'FS_StormDrain' AS tablename, MAX(version) FROM History_StormDrain;
-- SELECT 'FS_TimeCard' AS tablename, MAX(version) FROM History_TimeCard;
-- SELECT 'FS_TrapData' AS tablename, MAX(version) FROM History_TrapData;
-- SELECT 'FS_TrapLocation' AS tablename, MAX(version) FROM History_TrapLocation;
-- SELECT 'FS_Treatment' AS tablename, MAX(version) FROM History_Treatment;
-- SELECT 'FS_TreatmentArea' AS tablename, MAX(version) FROM History_TreatmentArea;
-- SELECT 'FS_Zones' AS tablename, MAX(version) FROM History_Zones;
-- SELECT 'FS_Zones2' AS tablename, MAX(version) FROM History_Zones2;

DROP FUNCTION IF EXISTS maxversion_all;
CREATE FUNCTION maxversion_all(schema_name text default 'public')
  RETURNS table(table_name text, max_version int) as
$$
declare
 table_name text;
begin
  for table_name in SELECT c.relname FROM pg_class c
    JOIN pg_namespace s ON (c.relnamespace=s.oid)
    WHERE c.relkind = 'r' AND s.nspname=schema_name AND c.relname LIKE 'history_%'
  LOOP
    RETURN QUERY EXECUTE format('select cast(%L as text),max(version) from %I.%I',
       table_name, schema_name, table_name);
  END LOOP;
end
$$ language plpgsql;
SELECT maxversion_all();
